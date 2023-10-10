FROM ghcr.io/initc3/linux-sgx:2.19-jammy as dev

RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys E5C7F0FA1C6C6C3C

RUN apt-get update && apt-get install \
    sudo \
    curl \
    git \
    clang \
    libclang-dev \
    apt-utils \
    build-essential \
    openssh-server \
    openjdk-11-jdk \
    -y

RUN mkdir /var/run/sshd
RUN echo 'root:root' | chpasswd

EXPOSE 22

# Create dev user with password 'dev'
RUN useradd -ms /bin/bash dev
RUN echo 'dev:dev' | chpasswd
RUN echo 'dev ALL=(ALL:ALL) ALL' >> /etc/sudoers

# RUN mkdir -p /home/dev/.ssh
# COPY ~/.ssh/id_rsa.pub /home/dev/.ssh/authorized_keys
# RUN chmod 600 /home/dev/.ssh/authorized_keys

RUN sudo -u dev bash -c "curl https://sh.rustup.rs -sSf | bash -s -- -y"
RUN bash -c "NONINTERACTIVE=1 $(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
RUN bash -c "(echo; echo 'eval \"$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)\"') >> /root/.profile"
RUN sudo -u dev bash -c "(echo; echo 'eval \"$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)\"') >> /home/dev/.profile"
RUN /home/linuxbrew/.linuxbrew/bin/brew install bazelisk go

# RUN chown -R dev:dev /home/dev/*
# RUN chown -R dev:dev /home/dev/.[^.]*

COPY /docker/dev/deventrypoint.sh /deventrypoint.sh
ENTRYPOINT ["/bin/sh", "/deventrypoint.sh"]

# Upon start, run ssh daemon
CMD ["/usr/sbin/sshd", "-D"]
