# Built by .github/workflows/deploy.yml (context ., file Dockerfile) and pushed
# to Artifact Registry.
#
# A job image, not a server: the default command runs the compiled Terratest
# suite, which really applies modules/pet-file with terraform, checks what it
# made, and destroys it. Exits 0 when every test passes. It never listens on $PORT.
#
# Two stages: the suite is compiled to a test binary with the Go toolchain, then
# run on the terraform image (the tests shell out to `terraform`). The providers
# the module needs are put in a plugin cache at build time.

FROM golang:1.27-alpine AS build
WORKDIR /src/test
COPY test/go.mod test/go.sum ./
RUN go mod download
COPY test/ ./
RUN CGO_ENABLED=0 go vet ./... \
 && CGO_ENABLED=0 go test -c -o /out/terratest.test .

FROM hashicorp/terraform:1.16.5 AS runtime
ARG BUILD_ID=""
ENV BUILD_ID=$BUILD_ID TF_IN_AUTOMATION=1 TF_INPUT=0 HOME=/home/app \
    TF_PLUGIN_CACHE_DIR=/home/app/.terraform.d/plugin-cache
RUN adduser -D -u 10001 -h /home/app app \
 && mkdir -p /app/test "$TF_PLUGIN_CACHE_DIR" && chown -R app:app /app /home/app
WORKDIR /app/test
COPY --chown=app:app modules /app/modules
COPY --from=build --chown=app:app /out/terratest.test /app/test/terratest.test
USER app
# warm the plugin cache from the module's lock file (a throwaway copy, so no .terraform ships)
RUN cp -r /app/modules/pet-file /tmp/warm \
 && terraform -chdir=/tmp/warm init -backend=false -input=false -lockfile=readonly \
 && rm -rf /tmp/warm
# the base image's ENTRYPOINT is `terraform`; the job is the test binary
ENTRYPOINT []
CMD ["/app/test/terratest.test", "-test.v", "-test.count=1", "-test.timeout=10m"]
