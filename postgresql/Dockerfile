FROM postgres:9.4

COPY docker-entrypoint-initdb.d/*.sh /docker-entrypoint-initdb.d/
RUN chmod 0755 /docker-entrypoint-initdb.d/*.sh

COPY docker-entrypoint-initdb.d/*.sql /docker-entrypoint-initdb.d/
