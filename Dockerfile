FROM alpine:3.11

RUN apk --no-cache add \
    curl \
    git \
    libc6-compat \
    openssh-client \
    python \
    bash \
    py-crcmod \
    py-pip && \
    pip install --upgrade pip==18.1

# Install a YAML Linter
ARG yamllint_version=1.21.0
LABEL yamllint_version=$yamllint_version
RUN pip install "yamllint==$yamllint_version"

# Install Yamale YAML schema validator
ARG yamale_version=2.0.1
LABEL yamale_version=$yamale_version
RUN pip install "yamale==$yamale_version"

# Install kubectl
ARG kubectl_version=v1.18.0
LABEL kubectl_version=$kubectl_version
RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$kubectl_version/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Install Helm
ARG helm_version=v3.2.4
LABEL helm_version=$helm_version
RUN curl -LO "https://get.helm.sh/helm-$helm_version-linux-amd64.tar.gz" && \
    mkdir -p "/usr/local/helm-$helm_version" && \
    tar -xzf "helm-$helm_version-linux-amd64.tar.gz" -C "/usr/local/helm-$helm_version" && \
    ln -s "/usr/local/helm-$helm_version/linux-amd64/helm" /usr/local/bin/helm && \
    rm -f "helm-$helm_version-linux-amd64.tar.gz"

COPY ./etc/chart_schema.yaml /etc/ct/chart_schema.yaml
COPY ./etc/lintconf.yaml /etc/ct/lintconf.yaml
COPY ct /usr/local/bin/ct
# Ensure that the binary is available on path and is executable
RUN ct --help
