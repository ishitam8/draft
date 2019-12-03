  
# Docker image file that describes an Ubuntu18.04 image with PowerShell installed from Microsoft APT Repo
ARG fromTag=18.04
ARG imageRepo=ubuntu

FROM ${imageRepo}:${fromTag} AS installer-env

ARG PS_VERSION=6.2.0
ARG PS_PACKAGE=powershell_${PS_VERSION}-1.ubuntu.18.04_amd64.deb
ARG PS_PACKAGE_URL=https://github.com/PowerShell/PowerShell/releases/download/v${PS_VERSION}/${PS_PACKAGE}

RUN echo ${PS_PACKAGE_URL}
# Download the Linux package and save it
ADD ${PS_PACKAGE_URL} /tmp/powershell.deb

# Define ENVs for Localization/Globalization
ENV DOTNET_SYSTEM_GLOBALIZATION_INVARIANT=false \
    LC_ALL=en_US.UTF-8 \
    LANG=en_US.UTF-8 \
    # set a fixed location for the Module analysis cache
    PSModuleAnalysisCachePath=/var/cache/microsoft/powershell/PSModuleAnalysisCache/ModuleAnalysisCache

# Install dependencies and clean up
RUN apt-get update \
    && apt-get install -y /tmp/powershell.deb \
    && apt-get install -y \
    # less is required for help in powershell
        less \
    # requied to setup the locale
        locales \
    # required for SSL
        ca-certificates \
        gss-ntlmssp \
        curl \
        apt-transport-https \
        lsb-release \
        gnupg \
        git \
    && apt-get dist-upgrade -y \
    && locale-gen $LANG && update-locale \
    # remove powershell package
    && rm /tmp/powershell.deb \
    # intialize powershell module cache
    && pwsh \
        -NoLogo \
        -NoProfile \
        -Command " \
          \$ErrorActionPreference = 'Stop' ; \
          \$ProgressPreference = 'SilentlyContinue' ; \
          while(!(Test-Path -Path \$env:PSModuleAnalysisCachePath)) {  \
            Write-Host "'Waiting for $env:PSModuleAnalysisCachePath'" ; \
            Start-Sleep -Seconds 6 ; \
          }" && \
    curl -sL https://packages.microsoft.com/keys/microsoft.asc | \
    gpg --dearmor | \
    tee /etc/apt/trusted.gpg.d/microsoft.asc.gpg > /dev/null && \ 
    export AZ_REPO=$(lsb_release -cs) && \
    echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ $AZ_REPO main" | \
    tee /etc/apt/sources.list.d/azure-cli.list && \ 
    apt-get update && \
    apt-get install azure-cli

# Define args needed only for the labels
ARG VCS_REF="none"
ARG IMAGE_NAME=powershell:ubuntu18.04
# Use PowerShell as the default shell
# Use array to avoid Docker prepending /bin/sh -c

#RUN apt-get update && apt-get install -y git curl 
RUN mkdir -p /draft/plugins
COPY ./_dist/linux-amd64/draft /bin
COPY ./docker /bin
ENV DRAFT_HOME=/draft
RUN ["draft", "init"]

CMD [ "pwsh" ]
