package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"log"
	"net/http"
	"net/url"
	"os"
)

type CepResponse struct {
	Localidade string `json:"localidade"`
	Uf         string `json:"uf"`
}

type ClimaResponse struct {
	Celsius    float64 `json:"celsius"`
	Fahrenheit float64 `json:"fahrenheit"`
	Kelvin     float64 `json:"kelvin"`
}

func initTracer() {
	//zipkinURL := "http://localhost:9411/api/v2/spans" // rodando da aplicação local
	zipkinURL := "http://zipkin:9411/api/v2/spans" // URL direta para o Zipkin

	exporter, err := zipkin.New(zipkinURL)
	if err != nil {
		log.Fatalf("failed to create Zipkin exporter | service B: %v", err)
	}

	serviceName := "service_b"
	bsp := trace.NewBatchSpanProcessor(exporter)

	tracerProvider := trace.NewTracerProvider(trace.WithSpanProcessor(bsp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)))
	otel.SetTracerProvider(tracerProvider)
}

func cepHandler(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("service_b")

	cep := r.URL.Query().Get("cep")

	_, span := tracer.Start(r.Context(), "tempo_de_execucao_clima")
	defer span.End()

	location, err := getLocationFromCep(cep)

	if err != nil {
		http.Error(w, "NÃO FOI POSSÍVEL LOCALIZAR CEP", http.StatusNotFound)
		return
	}

	weather, err := getWeather(location)
	if err != nil {
		http.Error(w, "FALHA AO TENTAR OBTER TEMPERATURA", http.StatusInternalServerError)
		return
	}

	response := ClimaResponse{
		Celsius:    weather.Celsius,
		Fahrenheit: weather.Celsius*1.8 + 32,
		Kelvin:     weather.Celsius + 273.15,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func getLocationFromCep(cep string) (CepResponse, error) {
	resp, err := http.Get("https://viacep.com.br/ws/" + cep + "/json/")
	if err != nil {
		log.Printf("Erro ao fazer a requisição para o ViaCEP: %v", err)

		return CepResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CepResponse{}, fmt.Errorf("invalid cep")
	}

	var location CepResponse
	err = json.NewDecoder(resp.Body).Decode(&location)
	return location, err
}

func getWeather(location CepResponse) (ClimaResponse, error) {
	query := url.QueryEscape(location.Localidade + "," + location.Uf)

	apiKey := os.Getenv("API_KEY")
	urlNew := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, query)

	resp, err := http.Get(urlNew)
	if err != nil {
		return ClimaResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ClimaResponse{}, fmt.Errorf("falha ao tentar pegar o clima")
	}

	var weather ClimaResponse
	err = json.NewDecoder(resp.Body).Decode(&weather)
	return weather, err
}

func main() {
	initTracer()
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		return
	}

	http.HandleFunc("/clima", cepHandler)
	fmt.Printf("Service B is running on port 8081")
	http.ListenAndServe(":8081", nil)
}
