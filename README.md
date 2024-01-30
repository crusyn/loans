# loan servicing application

## running the server

install go: https://go.dev/doc/install

```
go run main.go
```

The server will run at http://localhost:8080/

## api docs

You can play with the running api with Swagger Docs:
http://localhost:8080/swagger/index.html

## automated testing

```
go test ./...
```

## monthly payment calcuation

In any instance when rounding was required I opted to round up to be sure the bank is paid enough interest and principal.
In order to get the whole principal paid in the loan term I needed to add a penny to the monthly payment and then credited the aggregate overpayment in the last month.