# build go base
FROM golang:alpine AS go-builder
WORKDIR /app
COPY ./ ./
RUN go build -o /bin/server ./cmd/server

FROM node:alpine AS node-builder
WORKDIR /site
COPY ./site .
RUN npm install
ENV PUBLIC_API_URL=""
RUN npm run build

FROM scratch
COPY --from=go-builder /bin/server /bin/server
COPY --from=node-builder /site/build /public
ENV PUBLIC_DIR=/public

ENTRYPOINT ["/bin/server"]