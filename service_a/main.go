package main

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"io"
	"log"
	"net/http"
)

var validate = validator.New()

func initTracer() {
	//zipkinURL := "http://localhost:9411/api/v2/spans" // URL do Zipkin rodando fora do container
	zipkinURL := "http://zipkin:9411/api/v2/spans" // URL direta para o Zipkin

	exporter, err := zipkin.New(zipkinURL)
	if err != nil {
		log.Fatalf("failed to create Zipkin exporter: %v", err)
	}

	serviceName := "service_a"
	bsp := trace.NewBatchSpanProcessor(exporter)

	tracerProvider := trace.NewTracerProvider(trace.WithSpanProcessor(bsp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)))
	otel.SetTracerProvider(tracerProvider)
}

func cepHandler(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("service_a")
	ctx, span := tracer.Start(r.Context(), "tempo_de_execucao_cep")
	defer span.End()

	cep := r.URL.Query().Get("cep")

	if err := validate.Var(cep, "len=8,numeric"); err != nil {
		http.Error(w, "CEP INCORRETO", http.StatusUnprocessableEntity)
		return
	}

	callServiceB(ctx, cep, w)
}

func callServiceB(ctx context.Context, cep string, w http.ResponseWriter) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://service_b:8081/clima?cep="+cep, nil) // todo para rodar LOCAL
	//req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8081/clima?cep="+cep, nil)
	if err != nil {
		http.Error(w, "Erro ao chamar Serviço B", http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Erro ao chamar Serviço B - 2° parte: %v", err)
		http.Error(w, "Erro ao chamar Serviço B - 2° parte", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	for key, value := range resp.Header {
		w.Header()[key] = value
	}

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Erro ao copiar o corpo da resposta: %v", err)
		http.Error(w, "Erro ao copiar o corpo da resposta", http.StatusInternalServerError)
		return
	}
}

func main() {
	initTracer()
	http.HandleFunc("/cep", cepHandler)
	fmt.Printf("Service A is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
