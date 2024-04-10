FROM postgres:13

ENV POSTGRES_USER=pdao
ENV POSTGRES_PASSWORD=parallel

COPY init.sql /docker-entrypoint-initdb.d/

EXPOSE 5432
