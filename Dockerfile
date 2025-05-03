FROM debian:trixie-slim@sha256:1f87f0180766b28b7834a0b5f134948f9c2fea18ffa09bd62a7c93cc66ef5d99

WORKDIR /app

RUN apt-get update && apt-get install git python3 python3-venv golang ffmpeg -y
RUN python3 -m venv /app/venv
RUN /app/venv/bin/pip install "git+https://github.com/openai/whisper.git"

# cache go files
COPY go.* /app/
RUN go mod download

COPY . .
RUN go build

CMD ["./whisper_bot"]
