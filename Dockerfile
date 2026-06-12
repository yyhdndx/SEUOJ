FROM golang:1.25-alpine AS backend-build

WORKDIR /src/seu-oj-backend
COPY seu-oj-backend/go.mod seu-oj-backend/go.sum ./
RUN go mod download

COPY seu-oj-backend ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/seu-oj-web . \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/judge-worker ./cmd/judge-worker \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/db-init ./cmd/db-init

FROM node:22-alpine AS frontend-build

WORKDIR /src/seu-oj-frontend/CodeMirror
COPY seu-oj-frontend/CodeMirror/package.json seu-oj-frontend/CodeMirror/package-lock.json ./
RUN npm ci --omit=dev

WORKDIR /src/seu-oj-frontend
COPY seu-oj-frontend ./

FROM docker:27-cli AS runtime

WORKDIR /app/seu-oj-backend

COPY --from=backend-build /out/ /app/bin/
COPY seu-oj-backend/config ./config
COPY seu-oj-backend/database ./database
COPY --from=frontend-build /src/seu-oj-frontend /app/seu-oj-frontend

RUN mkdir -p /app/seu-oj-backend/logs /judge-work

EXPOSE 8080

CMD ["/app/bin/seu-oj-web"]
