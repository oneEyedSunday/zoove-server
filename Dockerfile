FROM golang:1.14

COPY . /app
WORKDIR /app
RUN go get ./
ENV PORT=8300
ENV GOOS linux
ENV DB_URL=postgres://omgaxiwlvnkpbs:3f5269f91782298c7480d21c745c69c5001bf14b25fb209cffda18025635df1b@ec2-54-246-85-151.eu-west-1.compute.amazonaws.com:5432/d317dve0k1sqiq
ENV DEEZER_APP_ID=422202
ENV DEEZER_APP_SECRET=e39a5133a1ab818926e848e5695e644c
ENV DEEZER_API_BASE=https://api.deezer.com
ENV DEEZER_REDIRECT_URI=https://4bb4df9ecc0f.ngrok.io/deezer/oauth
ENV DEEZER_AUTH_BASE=https://connect.deezer.com/oauth
ENV JWT_SECRET=ijeiuiengeivm29429r3im=egvv4v3tr
ENV SPOTIFY_CLIENT_ID=0888e2de0fdc43acba22bbabf00189ce
ENV SPOTIFY_CLIENT_SECRET=54da6c3422de4870bda7cb0689214c6c
ENV SPOTIFY_REDIRECT_URI=https://4bb4df9ecc0f.ngrok.io/spotify/oauth
ENV SPOTIFY_API_BASE=https://api.spotify.com
ENV SPOTIFY_AUTH_BASE=https://accounts.spotify.com
RUN go run github.com/prisma/prisma-client-go prefetch
RUN go run github.com/prisma/prisma-client-go generate
RUN go run github.com/prisma/prisma-client-go migrate up --experimental
RUN go run github.com/prisma/prisma-client-go migrate save --experimental --create-db --name "prod"
RUN go build -o app
EXPOSE 8300
ENTRYPOINT [ "./app" ]