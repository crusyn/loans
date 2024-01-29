package handlers

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/crusyn/loans/ent"
	"github.com/crusyn/loans/ent/user"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	Ent *ent.Client
}

type newUserRequest struct {
	Name    string `json:"name"`
	Social  string `json:"social"`
	Address string `json:"address"`
}

type newLoanRequest struct {
	Amount   float64 `json:"amount"`
	Rate     float64 `json:"rate"`
	Months   int     `json:"months"`
	Borrower int     `json:"borrowerID"`
}

type loanResponse struct {
	Amount float64 `json:"amount"`
	Rate   float64 `json:"rate"`
	Term   int     `json:"term"`
}

type loanMonthResponseItem struct {
	Month            int     `json:"month"`
	RemainingBalance float64 `json:"remainingBalance"`
	MonthlyPayment   float64 `json:"monthlyPayment"`
}

func (h Handler) CreateUser(ctx *gin.Context) {

	var newUser newUserRequest

	if err := ctx.BindJSON(&newUser); err != nil {
		log.Debug().Msgf("%v", err)
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "user input malformed",
		})
		return
	}

	socialExists, err := h.Ent.User.Query().Where(user.Social(newUser.Social)).Exist(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal error",
		})
		return
	}

	if socialExists {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "user with social security number already exists",
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
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"newUserID": u.ID,
	})

}

func (h Handler) CreateLoan(ctx *gin.Context) {

	var newLoan newLoanRequest

	if err := ctx.BindJSON(&newLoan); err != nil {
		log.Debug().Msgf("%v", err)
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "new loan input malformed",
		})
		return
	}

	if newLoan.Amount <= 0 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "loan amount must be positive",
		})
		return
	}

	if newLoan.Rate <= 0 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "rate must be positive",
		})
		return
	}

	if newLoan.Months <= 0 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "term must be positive",
		})
		return
	}

	userExists, err := h.Ent.User.Query().Where(user.ID(newLoan.Borrower)).Exist(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal error",
		})
		return
	}

	if !userExists {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "borrower doesn't exist",
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
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"newLoanID": l.ID,
	})
}

func (h Handler) GetLoan(ctx *gin.Context) {
	id := ctx.Param("id")

	i, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "id must be numeric",
		})
		return
	}

	l, err := h.Ent.Loan.Get(ctx, i)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "could not find loan",
		})
		return
	}

	ctx.JSON(http.StatusOK, loanResponse{
		Amount: float64(l.Amount) / 100,
		Rate:   l.Rate,
		Term:   l.Term,
	})
}

func (h Handler) GetLoanSchedule(ctx *gin.Context) {
	id := ctx.Param("id")

	i, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "id must be numeric",
		})
		return
	}

	l, err := h.Ent.Loan.Get(ctx, i)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "could not find loan",
		})
		return
	}

	months := []loanMonthResponseItem{}

	schedule, err := CreateAmortizationSchedule(float64(l.Amount), l.Rate, l.Term)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not return generate amortization schedule",
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
