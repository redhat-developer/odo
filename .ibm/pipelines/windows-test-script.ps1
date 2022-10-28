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
    Shout "kill previous odo sessions"
    taskkill /IM "odo.exe" /F

    Shout "Cloning Repo"
    git clone --depth 1 $REPO $BUILD_NUMBER
    Push-Location $BUILD_NUMBER

    Shout "Checking out PR #$GIT_PR_NUMBER"
    git fetch --depth 1 origin pull/${GIT_PR_NUMBER}/head:pr${GIT_PR_NUMBER}
    git checkout pr${GIT_PR_NUMBER}

    Shout "Setup ENV variables"
    mkdir bin 
    mkdir artifacts

    $PATH = [Environment]::GetEnvironmentVariable("PATH")
    $GOBIN="$(Get-Location)\bin"
    [Environment]::SetEnvironmentVariable("GOBIN", "$GOBIN")
    [Environment]::SetEnvironmentVariable("PATH", "$GOBIN;$PATH")

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
    [Environment]::SetEnvironmentVariable("TEST_EXEC_NODES", "$TEST_EXEC_NODES") 
    [Environment]::SetEnvironmentVariable("SKIP_USER_LOGIN_TESTS","true")
    [Environment]::SetEnvironmentVariable("SKIP_WELCOMING_MESSAGES","true")
    # Integration tests detecting key press when running DevSession are not working on Windows
    [Environment]::SetEnvironmentVariable("SKIP_KEY_PRESS","true")

    Shout "Login IBMcloud"
    ibmcloud login --apikey ${API_KEY}
    ibmcloud target -r eu-de
    ibmcloud oc cluster config -c ${CLUSTER_ID}

    Shout "Login Openshift"
    oc login -u apikey -p ${API_KEY} ${IBM_OPENSHIFT_ENDPOINT}
    Check-ExitCode $LASTEXITCODE

    Shout "Getting Devfile proxy address"
    $DEVFILE_PROXY=$(oc get svc -n devfile-proxy nginx -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    if ( $LASTEXITCODE -eq 0 )
    {
        Shout "Using Devfile proxy: $DEVFILE_PROXY"
        [Environment]::SetEnvironmentVariable("DEVFILE_PROXY", "$DEVFILE_PROXY")
    }

    Shout "Create Binary"
    make install 
    Shout "Running test"
    make test-integration-cluster           | tee -a  C:\Users\Administrator.ANSIBLE-TEST-VS\AppData\Local\Temp\$LOGFILE
    Check-ExitCode $LASTEXITCODE
    make test-e2e           | tee -a  C:\Users\Administrator.ANSIBLE-TEST-VS\AppData\Local\Temp\$LOGFILE
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
$TEST_EXEC_NODES=$args[7]
Shout "Args Recived"


# Run test
Run-Test

# Post test cleanup
Shout "Cleanup" 
Do-Cleanup