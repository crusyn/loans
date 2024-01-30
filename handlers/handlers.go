package handlers

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/crusyn/loans/ent"
	"github.com/crusyn/loans/ent/loan"
	"github.com/crusyn/loans/ent/sharedloan"
	"github.com/crusyn/loans/ent/user"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	Ent *ent.Client
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type newUserRequest struct {
	Name    string `json:"name"`
	Social  string `json:"social"`
	Address string `json:"address"`
}

type newUserResponse struct {
	UserId int `json:"newUserId"`
}

// @Summary Creates User
// @Schemes
// @Description Creates User given a `newUserRequest`
// @Accept json
// @Produce json
// @Param newUserRequest body newUserRequest true "New User Request"
// @Success 200 {object} newUserResponse
// @Router /user [post]
func (h Handler) CreateUser(ctx *gin.Context) {
	var newUser newUserRequest
	if err := ctx.BindJSON(&newUser); err != nil {
		log.Debug().Msgf("%v", err)
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "user input malformed",
		})
		return
	}

	socialExists, err := h.Ent.User.Query().Where(user.Social(newUser.Social)).Exist(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal error",
		})
		return
	}
	if socialExists {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "user with social security number already exists",
		})
		return
	}

	u, err := h.Ent.User.Create().
		SetName(newUser.Name).
		SetSocial(newUser.Social).
		SetAddress(newUser.Address).
		Save(ctx)
	if err != nil {
		log.Err(err)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse{
		UserId: u.ID,
	})

}

type newLoanRequest struct {
	Amount   float64 `json:"amount"`
	Rate     float64 `json:"rate"`
	Months   int     `json:"months"`
	Borrower int     `json:"borrowerID"`
}

type newLoanResponse struct {
	LoanId int `json:"newLoanId"`
}

// @Summary Creates Loan
// @Schemes
// @Description Creates a Loan associated with a specific borrower
// @Accept json
// @Produce json
// @Param newLoanRequest body newLoanRequest true "New Loan Request"
// @Success 200 {object} newLoanResponse
// @Router /loan/ [post]
func (h Handler) CreateLoan(ctx *gin.Context) {
	var newLoan newLoanRequest
	if err := ctx.BindJSON(&newLoan); err != nil {
		log.Debug().Msgf("%v", err)
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "new loan input malformed",
		})
		return
	}

	if newLoan.Amount <= 0 {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "loan amount must be positive",
		})
		return
	}
	if newLoan.Rate <= 0 {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "rate must be positive",
		})
		return
	}
	if newLoan.Months <= 0 {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "term must be positive",
		})
		return
	}

	userExists, err := h.Ent.User.Query().Where(user.ID(newLoan.Borrower)).Exist(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal error",
		})
		return
	}
	if !userExists {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "borrower doesn't exist",
		})
		return
	}

	l, err := h.Ent.Loan.Create().
		SetAmount(int(newLoan.Amount * 100)).
		SetRate(newLoan.Rate).
		SetTerm(newLoan.Months).
		SetBorrowerID(newLoan.Borrower).
		Save(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, newLoanResponse{
		LoanId: l.ID,
	})
}

type loanResponse struct {
	Id     int     `json:"id"`
	Amount float64 `json:"amount"`
	Rate   float64 `json:"rate"`
	Term   int     `json:"term"`
}

// @Summary Gets Loan Information
// @Schemes
// @Description Gets Loan Terms
// @Accept json
// @Produce json
// @Param loanid path int true "Loan Id"
// @Success 200 {object} loanResponse
// @Router /loan/{loanid} [get]
func (h Handler) GetLoan(ctx *gin.Context) {
	id := ctx.Param("id")

	i, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "id must be numeric",
		})
		return
	}

	l, err := h.Ent.Loan.Get(ctx, i)
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Message: "could not find loan",
		})
		return
	}

	ctx.JSON(http.StatusOK, loanResponse{
		Id:     l.ID,
		Amount: float64(l.Amount) / 100,
		Rate:   l.Rate,
		Term:   l.Term,
	})
}

// @Summary Gets Loans by User
// @Schemes
// @Description Gets Loans associated with a specific user.  The user may be the borrower
// @Description or the loan may be shared with that user.
// @Accept json
// @Produce json
// @Param userid path int true "User Id"
// @Success 200 {array} loanResponse
// @Router /user/{userid}/loans [get]
func (h Handler) GetLoans(ctx *gin.Context) {
	id := ctx.Param("id")

	userId, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "id must be numeric",
		})
		return
	}

	userExists, err := h.Ent.User.Query().Where(user.ID(userId)).Exist(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal error",
		})
		return
	}

	if !userExists {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "user doesn't exist",
		})
		return
	}

	loans, err := h.Ent.Loan.Query().
		Where(loan.BorrowerID(userId)).
		All(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal error",
		})
		return
	}

	response := []loanResponse{}

	for _, l := range loans {
		response = append(response, loanResponse{
			Id:     l.ID,
			Amount: float64(l.Amount) / 100,
			Rate:   l.Rate,
			Term:   l.Term,
		})
	}

	sharedLoans, err := h.Ent.SharedLoan.Query().
		Where(sharedloan.UserID(userId)).
		WithLoan().
		All(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal error",
		})
		return
	}

	for _, l := range sharedLoans {
		response = append(response, loanResponse{
			Id:     l.Edges.Loan.ID,
			Amount: float64(l.Edges.Loan.Amount) / 100,
			Rate:   l.Edges.Loan.Rate,
			Term:   l.Edges.Loan.Term,
		})
	}

	ctx.JSON(http.StatusOK, response)
}

type loanMonthResponseItem struct {
	Month            int     `json:"month"`
	RemainingBalance float64 `json:"remainingBalance"`
	MonthlyPayment   float64 `json:"monthlyPayment"`
}

// @Summary Gets Loan Schedule
// @Schemes
// @Description Gets the loans schedule by month
// @Accept json
// @Produce json
// @Param loanid path int true "Loan Id"
// @Success 200 {array} loanMonthResponseItem
// @Router /loan/{loanid}/schedule [get]
func (h Handler) GetLoanSchedule(ctx *gin.Context) {
	id := ctx.Param("id")

	i, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "id must be numeric",
		})
		return
	}

	l, err := h.Ent.Loan.Get(ctx, i)
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Message: "could not find loan",
		})
		return
	}

	months := []loanMonthResponseItem{}

	schedule, err := CreateAmortizationSchedule(float64(l.Amount)/100, l.Rate, l.Term)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "could not generate amortization schedule",
		})
		return
	}

	for _, m := range schedule {
		months = append(months, loanMonthResponseItem{
			Month:            m.Month,
			RemainingBalance: m.EndingBalance,
			MonthlyPayment:   m.MonthlyPayment,
		})
	}

	ctx.JSON(http.StatusOK, months)

}

type loanMonthSummaryResponse struct {
	EndingBalance      float64 `json:"endingBalance"`
	TotalPrincipalPaid float64 `json:"totalPrincipalPaid"`
	TotalInterestPaid  float64 `json:"totalInterestPaid"`
}

// @Summary Gets Loan Month Summary
// @Schemes
// @Description Gets aggregate loan data given a particular month
// @Accept json
// @Produce json
// @Param loanid path int true "Loan Id"
// @Param month path int true "Month Number"
// @Success 200 {object} loanMonthSummaryResponse
// @Router /loan/{loanid}/month/{month} [get]
func (h Handler) GetMonthSummary(ctx *gin.Context) {
	id := ctx.Param("id")

	i, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "id must be numeric",
		})
		return
	}

	l, err := h.Ent.Loan.Get(ctx, i)
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Message: "could not find loan",
		})
		return
	}

	schedule, err := CreateAmortizationSchedule(float64(l.Amount)/100, l.Rate, l.Term)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "could not generate amortization schedule",
		})
		return
	}

	month := ctx.Param("number")
	n, err := strconv.Atoi(month)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "month number must be numeric",
		})
		return
	}

	if n > l.Term {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "month number cannot be greater than term",
		})
		return
	}

	ctx.JSON(http.StatusOK, loanMonthSummaryResponse{
		EndingBalance:      schedule[n-1].EndingBalance,
		TotalPrincipalPaid: schedule[n-1].TotalPrincipalPaid,
		TotalInterestPaid:  schedule[n-1].TotalInterestPaid,
	})
}

type loanShareRequest struct {
	UserId int `json:"id"`
}

// @Summary Shares Loan
// @Schemes
// @Description Shares loan with another user that is not the borrower
// @Accept json
// @Produce json
// @Param loanid path int true "Loan Id"
// @Param loanShareRequest body loanShareRequest true "Share User Request"
// @Success 200
// @Router /loan/{loanid}/share [post]
func (h Handler) ShareLoan(ctx *gin.Context) {
	id := ctx.Param("id")

	loanId, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "loan id must be numeric",
		})
		return
	}

	loanExists, err := h.Ent.Loan.Query().
		Where(loan.ID(loanId)).
		Exist(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal error",
		})
		return
	}

	if !loanExists {
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Message: "could not find loan",
		})
		return
	}

	l, err := h.Ent.Loan.Get(ctx, loanId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal error",
		})
		return
	}

	var req loanShareRequest

	if err := ctx.BindJSON(&req); err != nil {
		log.Debug().Msgf("%v", err)
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "share request input malformed",
		})
		return
	}

	if l.BorrowerID == req.UserId {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "cannot share with borrower",
		})
		return
	}

	userExists, err := h.Ent.User.Query().Where(user.ID(req.UserId)).Exist(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal error",
		})
		return
	}

	if !userExists {
		ctx.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Message: "user doesn't exist",
		})
		return
	}

	loanShareExists, err := h.Ent.SharedLoan.Query().Where(
		sharedloan.And(
			sharedloan.UserID(req.UserId),
			sharedloan.LoanID(loanId),
		)).Exist(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal error",
		})
		return
	}

	if loanShareExists {
		ctx.JSON(http.StatusOK, ErrorResponse{
			Message: "shared loan already exists",
		})
		return
	}

	err = h.Ent.SharedLoan.Create().
		SetLoanID(loanId).
		SetUserID(req.UserId).
		Exec(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "internal error",
		})
		return
	}
}

func monthlyPayment(loanAmountCents int, annualInterestRate float64, termMonths int) (int, error) {
	// calculated using https://www.investopedia.com/terms/a/amortization.asp formula

	if loanAmountCents <= 0 {
		return 0, errors.New("loan amount must be positive")
	}
	if annualInterestRate <= 0 {
		return 0, errors.New("interest rate must be positive")
	}
	if termMonths <= 0 {
		return 0, errors.New("number of payments must be positive")
	}

	monthlyInterestRate := annualInterestRate / 12
	compoundedInterest := math.Pow(1+monthlyInterestRate, float64(termMonths))
	factor := (monthlyInterestRate * compoundedInterest) / (compoundedInterest - 1)
	return int(math.Ceil(float64(loanAmountCents)*factor)) + 1, nil
}

type monthlySummary struct {
	Month              int
	BeginningBalance   float64
	EndingBalance      float64
	MonthlyPayment     float64
	TotalPrincipalPaid float64
	TotalInterestPaid  float64
	CurrentInterest    float64
	CurrentPrincipal   float64
}

func CreateAmortizationSchedule(loanAmount float64, annualInterestRate float64, termMonths int) ([]monthlySummary, error) {
	loanAmountCents := int(loanAmount * 100)

	paymentCents, err := monthlyPayment(loanAmountCents, annualInterestRate, termMonths)
	if err != nil {
		return nil, err
	}

	summaries := make([]monthlySummary, termMonths)

	outstandingBeginningBalance := loanAmountCents
	totalPricipalPaid := 0
	totalInterestPaid := 0
	i := 0
	for i < termMonths {
		currentInterest := int(math.Ceil(float64(outstandingBeginningBalance) * (annualInterestRate / 12)))
		currentPrinciple := paymentCents - currentInterest
		if outstandingBeginningBalance < currentPrinciple {
			currentPrinciple = outstandingBeginningBalance
		}
		totalInterestPaid = totalInterestPaid + currentInterest
		totalPricipalPaid = totalPricipalPaid + currentPrinciple
		endingBalance := outstandingBeginningBalance - currentPrinciple

		summaries[i] = monthlySummary{
			Month:              i + 1,
			BeginningBalance:   float64(outstandingBeginningBalance) / 100,
			MonthlyPayment:     float64(currentInterest+currentPrinciple) / 100,
			CurrentInterest:    float64(currentInterest) / 100,
			CurrentPrincipal:   float64(currentPrinciple) / 100,
			TotalPrincipalPaid: float64(totalPricipalPaid) / 100,
			TotalInterestPaid:  float64(totalInterestPaid) / 100,
			EndingBalance:      float64(endingBalance) / 100,
		}

		outstandingBeginningBalance = endingBalance
		i = i + 1
	}

	return summaries, nil
}
