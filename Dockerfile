FROM gcr.io/distroless/static-debian12

COPY gobl.html /

ENTRYPOINT ["/gobl.html"]

EXPOSE 3000