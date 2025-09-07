# imag2webp
建议使用Linux环境进行运行，win平台cgo支持不是很好

# how to use?
```shell
$ make help
Available commands:
  make build          - Build the Go binary
  make run            - Run the application locally
  make docker-build   - Build Docker image with default settings
  make docker-build-env - Build Docker image using .env variables
  make docker-run     - Run Docker container

# run in docker
make docker-run

# test
$ curl -X POST http://localhost:10080/v1/upload   -F "image=@test.png"
Warning: Binary output can mess up your terminal. Use "--output -" to tell 
Warning: curl to output it to your terminal anyway, or consider "--output 
Warning: <FILE>" to save to a file.
```
