FROM golang:1.9.1
WORKDIR ~/go/src/github.com/reinho/cdp-screenshots
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -v

FROM ubuntu:16.04
RUN apt-get update && apt-get install -y wget && rm -r /var/lib/apt/lists/*
RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
RUN echo 'deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main' | tee /etc/apt/sources.list.d/google-chrome.list
RUN apt-get update && apt-get install -y google-chrome-unstable && rm -r /var/lib/apt/lists/*
RUN groupadd -r chrome && useradd -r -g chrome chrome && mkdir /home/chrome && chown -R chrome:chrome /home/chrome
COPY --from=0 ~/go/src/github.com/reinho/cdp-screenshots/cdp-screenshots /usr/bin

USER google-chrome
ENTRYPOINT ["cdp-screenshots"]
