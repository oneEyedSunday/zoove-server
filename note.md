dotenv -e .env.dev -- go run github.com/prisma/prisma-client-go generate
dotenv -e .env.dev -- go run github.com/prisma/prisma-client-go migrate save --experimental --create-db --name "third"
dotenv -e .env.dev -- go run github.com/prisma/prisma-client-go migrate up --experimental

