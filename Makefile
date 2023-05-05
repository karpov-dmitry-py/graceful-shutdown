run: .run ## run app
.PHONY: .run
.run:
	$(info running app ...)
	go run cmd/main.go


build: .build ## build an executable app
.PHONY: .build
.build:
	$(info building app ...)
	go build -o bin/app cmd/main.go