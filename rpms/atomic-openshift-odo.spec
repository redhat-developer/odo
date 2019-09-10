#this is a template spec and actual spec will be generated
#debuginfo not supported with Go
%global debug_package %{nil}
%global package_name atomic-openshift-odo
%global product_name odo
%global golang_version 1.11
%global odo_version ${ODO_RPM_VERSION}
%global odo_release ${ODO_RELEASE}
%global git_commit  ${GIT_COMMIT}
%global odo_cli_version v%{odo_version}
%global source_dir atomic-openshift-odo-%{odo_version}-%{odo_release}
%global source_tar %{source_dir}.tar.gz
%global gopath  %{_builddir}/gocode

Name:           %{package_name}
Version:        %{odo_version}
Release:        %{odo_release}%{?dist}
Summary:        %{product_name} client odo CLI binary
License:        ASL 2.0
URL:            https://github.com/openshift/odo/tree/%{odo_cli_version}

ExclusiveArch:  x86_64

Source0:        %{source_tar}
BuildRequires:  gcc
BuildRequires:  golang >= %{golang_version}
Provides:       %{package_name}
Obsoletes:      %{package_name}

%description
OpenShift Do (odo) is a fast, iterative, and straightforward CLI tool for developers who write, build, and deploy applications on OpenShift.

%prep
%setup -q -n %{source_dir}

%build
export GITCOMMIT="%{git_commit}"
mkdir -p %{gopath}/src/github.com/openshift
ln -s "$(pwd)" %{gopath}/src/github.com/openshift/odo
export GOPATH=%{gopath}
cd %{gopath}/src/github.com/openshift/odo
make cross
unlink %{gopath}/src/github.com/openshift/odo
rm -rf %{gopath}

%install
mkdir -p %{buildroot}/%{_bindir}
install -m 0755 dist/bin/linux-amd64/odo %{buildroot}/%{_bindir}/odo
mkdir -p %{buildroot}%{_datadir}
install -d %{buildroot}%{_datadir}/%{name}/{linux,macos,windows}
install -p -m 755 dist/bin/linux-amd64/odo %{buildroot}%{_datadir}/%{name}/linux/odo-linux-amd64
install -p -m 755 dist/bin/darwin-amd64/odo %{buildroot}%{_datadir}/%{name}/macos/odo-darwin-amd64
install -p -m 755 dist/bin/windows-amd64/odo.exe %{buildroot}%{_datadir}/%{name}/windows/odo-windows-amd64.exe

%files
%license LICENSE
%{_bindir}/odo

%package redistributable
Summary:        %{product_name} client CLI binaries for Linux, macOS and Windows
BuildRequires:  gcc
BuildRequires:  golang >= %{golang_version}
Provides:       %{package_name}-redistributable
Obsoletes:      %{package_name}-redistributable

%description redistributable
%{product_name} client odo cross platform binaries for Linux, macOS and Windows.

%files redistributable
%license LICENSE
%dir %{_datadir}/%{name}/linux/
%dir %{_datadir}/%{name}/macos/
%dir %{_datadir}/%{name}/windows/
%{_datadir}/%{name}/linux/odo-linux-amd64
%{_datadir}/%{name}/macos/odo-darwin-amd64
%{_datadir}/%{name}/windows/odo-windows-amd64.exe

