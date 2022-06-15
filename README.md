
# Usage

Start the postgres database:
```bash
cd database
docker build . -t rest-ws
docker run -d -p 54321:5432 rest-ws
``` 
Modify the database url in `.env.example` as necessary and then run:

```bash
cp .env.example .env
cd ..
```

Now you can execute the package as follows:
```bash
go run .
```
or
```bash
go build
./rest-ws
```


