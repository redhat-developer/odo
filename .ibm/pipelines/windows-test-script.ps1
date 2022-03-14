function Shout {
    Write-Host "--------------------------------------------------------------------------------"
    Write-Host $args[0]
    Write-Host "--------------------------------------------------------------------------------"
}

function Check-ExitCode {
    if ($args[0] -ne 0) {
        Write-Host "ERROR: Script returned a non-zero exit code"
        Pop-Location 
        Do-Cleanup
        exit 1
    }
}

function Do-Cleanup {
    if (Test-Path -Path  $BUILD_NUMBER) {  
    Shout "Removed $BUILD_NUMBER"
    Remove-Item  -Force -Recurse  $BUILD_NUMBER
    }
}    

function Run-Test {
    Shout "Cloning Repo"
    git clone $REPO $BUILD_NUMBER
    Push-Location $BUILD_NUMBER

    Shout "Checkout to $GIT_PR_NUMBER"
    git fetch -v origin pull/${GIT_PR_NUMBER}/head:pr${GIT_PR_NUMBER}
    git checkout main
    git merge pr${GIT_PR_NUMBER} --no-edit

    Shout "Setup ENV variables"
    mkdir bin 
    mkdir artifacts

    $PATH = [Environment]::GetEnvironmentVariable("PATH")
    $GOBIN="$(Get-Location)\bin"
    [Environment]::SetEnvironmentVariable("GOBIN", "$GOBIN")
    [Environment]::SetEnvironmentVariable("PATH", "$PATH;$GOBIN")

    # Set kubeconfig to current dir. This ensures no clashes with other test runs
    [Environment]::SetEnvironmentVariable("KUBECONFIG","$(Get-Location)\config")
    $ARTIFACT_DIR=${ARTIFACT_DIR:-"$(Get-Location)\artifacts"}
    $CUSTOM_HOMEDIR=$ARTIFACT_DIR
    $WORKDIR=$(Get-Location)

    [Environment]::SetEnvironmentVariable("ARTIFACT_DIR",${ARTIFACT_DIR:-"$(pwd)\artifacts"})
    [Environment]::SetEnvironmentVariable("CUSTOM_HOMEDIR",$ARTIFACT_DIR)
    [Environment]::SetEnvironmentVariable("WORKDIR",${WORKDIR:-"$(pwd)"})

    $GOCACHE="$(Get-Location)\.gocache" 
    mkdir $GOCACHE
    [Environment]::SetEnvironmentVariable("GOCACHE", "$GOCACHE")

    [Environment]::SetEnvironmentVariable("SKIP_USER_LOGIN_TESTS","true")

    Shout "Login IBMcloud"
    ibmcloud login --apikey ${API_KEY}
    ibmcloud target -r eu-de
    ibmcloud oc cluster config -c ${CLUSTER_ID}

    Shout "Login Openshift"
    oc login -u apikey -p ${API_KEY} ${IBM_OPENSHIFT_ENDPOINT}
    Check-ExitCode $LASTEXITCODE

    Shout "Create Binary"
    make install 
    Shout "Running test"
    make test-integration-devfile   | tee -a  C:\Users\Administrator.ANSIBLE-TEST-VS\AppData\Local\Temp\$LOGFILE
    Check-ExitCode $LASTEXITCODE
    make test-integration           | tee -a  C:\Users\Administrator.ANSIBLE-TEST-VS\AppData\Local\Temp\$LOGFILE
    Check-ExitCode $LASTEXITCODE
    make test-cmd-login-logout      | tee -a  C:\Users\Administrator.ANSIBLE-TEST-VS\AppData\Local\Temp\$LOGFILE
    Check-ExitCode $LASTEXITCODE
    make test-cmd-project           | tee -a  C:\Users\Administrator.ANSIBLE-TEST-VS\AppData\Local\Temp\$LOGFILE
    Check-ExitCode $LASTEXITCODE
    make test-e2e-devfile           | tee -a  C:\Users\Administrator.ANSIBLE-TEST-VS\AppData\Local\Temp\$LOGFILE
    Check-ExitCode $LASTEXITCODE

    Pop-Location 
    Shout "Test Finished"
}

$GIT_PR_NUMBER=$args[0]
$BUILD_NUMBER=$args[1]
$API_KEY=$args[2]
$IBM_OPENSHIFT_ENDPOINT=$args[3]
$LOGFILE=$args[4]
$REPO=$args[5]
$CLUSTER_ID=$args[6]
Shout "Args Recived"

# Pre test cleanup
Do-Cleanup

# Run test
Run-Test

# Post test cleanup
Shout "Cleanup" 
Do-Cleanup