# build go base
FROM golang:alpine AS go-builder
WORKDIR /app
COPY ./ ./
RUN go build -o /bin/server ./cmd/http-server

# run npm run build
FROM node:alpine AS node-builder
WORKDIR /site
COPY ./site .
RUN npm install
RUN npm run build

FROM scratch
COPY --from=go-builder /bin/server /bin/server
COPY --from=node-builder /site/dist /public
ENV PUBLIC_DIR=/public

ENTRYPOINT ["/bin/server"]