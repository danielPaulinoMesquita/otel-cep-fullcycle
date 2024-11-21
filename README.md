## Boa tarde, pessoal!

Segue o desafio prático de Observabilidade & Open Telemetry.


# Executar o projeto.

Você deve rodar o docker compose para testar aplicação, então rode o comando para:

### Criar Imagem: 
* docker compose build --no-cache

### Verifique se as imagens foram criadas:
* docker images

### Executar Serviços:
* docker compose up -d

## Chamar a API: 
http://localhost:8080/cep?cep=SEU_CEP_AQUI


## Ver logs
* Nome do span do tempo de execução do service A é: tempo_de_execucao_clima
* Nome do span do tempo de execução do service B é: tempo_de_execucao_cep

http://127.0.0.1:9411/

## Testar pelo test.http
Existe também um arquivo pronta para fazer fazer as chamadas caso não queira usar o postman. Ele fica em:
* cd test-services/test.http


## Rodar localmente:
Rodar o Serviço A:
* cd /service_a
Execute o comando:
* go run main.go
-----------

Rodar o Serviço B:
* cd /servico_b
Execute o comando:
* go run main.go
-------------

Rodar o zipkin
docker run -d -p 9411:9411 openzipkin/zipkin

Verificar os logs:
http://localhost:9411
ou
http://127.0.0.1:9411/