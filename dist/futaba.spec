%global commit 69d0c53515a91089692ccf08f0518098c2890fdb

Name:           futaba
Version:        1.0.0
Release:        1%{?dist}
Summary:        A silly Discord bot for a friend.

License:        GPL-3.0-or-later
URL:            https://gitlab.com/losuler/%{name}
Source0:        https://gitlab.com/losuler/%{name}/-/archive/%{commit}/%{name}-%{commit}.tar.gz

BuildRequires:  golang >= 1.12

%description
Futaba is a silly Discord bot for a friend. It has the following features:

- Show the current time of a user on your server.

%prep
%setup -q -n %{name}-%{commit}

%build
go build \
    -ldflags "${LDFLAGS:-} -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \n')%{?__global_ldflags: -extldflags '%__global_ldflags'}" \
    -a -v

%install
install -D -m 0755 %{name} %{buildroot}/%{_bindir}/%{name}
install -D -m 0644 config.yml.example %{buildroot}/%{_sysconfdir}/%{name}.yml
install -D -m 0644 dist/%{name}.service %{buildroot}/%{_unitdir}/%{name}.service

%files
%license LICENSE.txt
%{_bindir}/%{name}
%config %{_sysconfdir}/%{name}.yml
%{_unitdir}/%{name}.service

%changelog
