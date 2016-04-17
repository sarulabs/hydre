FROM php:7-fpm-alpine

RUN apk add --update nginx && rm -rf /var/cache/apk/*

COPY app /home/app
COPY nginx.conf /etc/nginx/nginx.conf
COPY hydre.yml /home/hydre.yml
ADD https://github.com/sarulabs/hydre/releases/download/2.0.0/hydre /home/hydre
RUN chmod +x /home/hydre

EXPOSE 80

CMD ["/home/hydre", "-c", "/home/hydre.yml"]
