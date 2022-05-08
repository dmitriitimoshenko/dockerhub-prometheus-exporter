FROM golang 
 
WORKDIR /app 
 
COPY . .

EXPOSE 2113
 
ENTRYPOINT ["go", "run", "main.go"]