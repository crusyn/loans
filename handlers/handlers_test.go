package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"

	"github.com/crusyn/loans/ent"
	"github.com/crusyn/loans/ent/user"
	"github.com/gin-gonic/gin"
)

// mock gin context
func GetTestGinContext(w *httptest.ResponseRecorder) *gin.Context {
	gin.SetMode(gin.TestMode)

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = &http.Request{
		Header: make(http.Header),
		URL:    &url.URL{},
	}

	return ctx
}

func TestCreateUser(t *testing.T) {

	// db init
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatal().Msgf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatal().Msgf("failed creating schema resources: %v", err)
	}

	h := Handler{
		Ent: client,
	}

	for _, tc := range []struct {
		name         string
		request      newUserRequest
		expectedCode int
	}{
		{
			name: "first chris",
			request: newUserRequest{
				Name:    "chris",
				Social:  "123-45-6789",
				Address: "1 Apple Street",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "duplicate social",
			request: newUserRequest{
				Name:    "chris",
				Social:  "123-45-6789",
				Address: "1 Apple Street",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "duplicate name",
			request: newUserRequest{
				Name:    "chris",
				Social:  "000-45-6780",
				Address: "1 Apple Street",
			},
			expectedCode: http.StatusOK,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx := GetTestGinContext(w)

			ctx.Request.Method = "POST"
			ctx.Request.Header.Set("Content-Type", "application/json")

			jsonbytes, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatal("could not marshal user")
			}

			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(jsonbytes))

			h.CreateUser(ctx)

			socialExists, err := h.Ent.User.Query().Where(user.Social(tc.request.Social)).Exist(ctx)
			if err != nil {
				t.Fatalf("could not get user: %v", err)
			}

			if !socialExists {
				t.Errorf("user with social %s not saved", tc.request.Social)
			}

			if w.Code != tc.expectedCode {
				t.Errorf("unexpected status code, want: %v, got: %v", tc.expectedCode, w.Code)
			}
		})
	}

}

func TestAmortizationSchedule(t *testing.T) {
	for _, tc := range []struct {
		name        string
		loan        loanResponse
		monthNumber int
		summary     monthlySummary
	}{
		{
			name: "$1M @ 5% last month",
			loan: loanResponse{
				Amount: 1000000.0,
				Rate:   0.05,
				Term:   360,
			},
			monthNumber: 360,
			summary: monthlySummary{
				Month:              360,
				BeginningBalance:   5338.68,
				EndingBalance:      0,
				MonthlyPayment:     5360.93,
				TotalPrincipalPaid: 1000000,
				TotalInterestPaid:  932555.5,
				CurrentInterest:    22.25,
				CurrentPrincipal:   5338.68,
			},
		}, {
			name: "$1M @ 5% first month",
			loan: loanResponse{
				Amount: 1000000.0,
				Rate:   0.05,
				Term:   360,
			},
			monthNumber: 1,
			summary: monthlySummary{
				Month:              1,
				BeginningBalance:   1000000,
				EndingBalance:      998798.44,
				MonthlyPayment:     5368.23, // we add $0.01 to the monthly payment to make sure the principal is fully paid off
				TotalPrincipalPaid: 1201.56,
				TotalInterestPaid:  4166.67,
				CurrentInterest:    4166.67,
				CurrentPrincipal:   1201.56,
			},
		}, {
			name: "$1M @ 5% middle month",
			loan: loanResponse{
				Amount: 1000000.0,
				Rate:   0.05,
				Term:   360,
			},
			monthNumber: 158,
			summary: monthlySummary{
				Month:              158,
				BeginningBalance:   734428.79,
				EndingBalance:      732120.68,
				MonthlyPayment:     5368.23, // we add $0.01 to the monthly payment to make sure the principal is fully paid off
				TotalPrincipalPaid: 267879.32,
				TotalInterestPaid:  580301.02,
				CurrentInterest:    3060.12,
				CurrentPrincipal:   2308.11,
			},
		}, {
			name: "$1.2M @ 11.5% first month",
			loan: loanResponse{
				Amount: 1212530.0,
				Rate:   0.115,
				Term:   360,
			},
			monthNumber: 1,
			summary: monthlySummary{
				Month:              1,
				BeginningBalance:   1212530.0,
				EndingBalance:      1212142.48,
				MonthlyPayment:     12007.60, // we add $0.01 to the monthly payment to make sure the principal is fully paid off
				TotalPrincipalPaid: 387.52,
				TotalInterestPaid:  11620.08,
				CurrentInterest:    11620.08,
				CurrentPrincipal:   387.52,
			},
		}, {
			name: "$1.2M @ 11.5% last month",
			loan: loanResponse{
				Amount: 1212530.0,
				Rate:   0.115,
				Term:   360,
			},
			monthNumber: 360,
			summary: monthlySummary{
				Month:              360,
				BeginningBalance:   11849.11,
				EndingBalance:      0,
				MonthlyPayment:     11962.67,
				TotalPrincipalPaid: 1212530.0,
				TotalInterestPaid:  3110161.07,
				CurrentInterest:    113.56,
				CurrentPrincipal:   11849.11,
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			amortizationSchedule, err := CreateAmortizationSchedule(tc.loan.Amount, tc.loan.Rate, tc.loan.Term)
			if err != nil {
				t.Fatalf("could not create amortization schedule: %v", err)
			}
			if diff := cmp.Diff(tc.summary, amortizationSchedule[tc.monthNumber-1]); diff != "" {
				t.Errorf("unexpected summary for month %d, (-want +got) %s", tc.monthNumber, diff)
			}
		})
	}
}
