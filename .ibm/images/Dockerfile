FROM golang:1.17

RUN curl -fsSL https://clis.cloud.ibm.com/install/linux | sh && \
    curl -sLO https://github.com/cli/cli/releases/download/v2.1.0/gh_2.1.0_linux_amd64.deb && \
    apt install ./gh_2.1.0_linux_amd64.deb && \
    curl -sLO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/ && \
    curl -sLO https://raw.githubusercontent.com/cptmorgan-rh/install-oc-tools/master/install-oc-tools.sh > /dev/null && \
    chmod +x install-oc-tools.sh && \
    ./install-oc-tools.sh --latest 4.7 && \
    apt-get update -y && \
    apt-get install -y sshpass && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user and associated home directory
RUN useradd -u 2001 --create-home tester
# Change to non-root privilege
USER tester

RUN go get github.com/kadel/odo-robot@965ea0dd848856691bfc76e6824a8b787b950045 && \
    ibmcloud plugin install -f cloud-object-storage && \
    ibmcloud plugin install -f kubernetes-service
