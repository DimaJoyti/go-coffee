#!/bin/bash
cd /home/dima/Desktop/Fun/Projects/go-coffee/crypto-terminal
go get go.opentelemetry.io/otel/sdk@latest
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@latest
go get go.opentelemetry.io/otel/exporters/prometheus@latest
go mod tidy
