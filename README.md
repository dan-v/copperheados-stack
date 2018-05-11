## What is CopperheadOS
CopperheadOS is a Linux-based mobile operating system with a focus on privacy and security. It builds on the latest stable release of the Android Open Source Project which is Android without any Google apps and services.

## What is copperheados-stack
<b>copperheados-stack</b> is an tool that will deploy all the AWS infrastructure required to run your own [CopperheadOS](https://copperhead.co/android/) build and release environment. It uses AWS Lambda to check for new releases, provisions EC2 spot instances for OS builds on demand, and uploads build artifacts to S3. Resulting OS builds are configured to receive over the air updates from this environment.

## Features
* Support for Google Pixel, Pixel XL, Pixel 2, and Pixel 2 XL
* End to end setup of build environment for CopperheadOS in AWS
* OTA updates through built in updater app - no need to manually flash your device on each new release
* Scheduled Lambda function looks for new releases to build on a daily basis
* Costs a few dollars a month to run (EC2 spot instance and S3 storage costs)

## Supporting CopperheadOS
If you use this tool, I <b>HIGHLY</b> recommend supporting the project with donations: https://copperhead.co/android/donate. 

## Installation
The easiest way is to download a pre-built binary from the [GitHub Releases](https://github.com/dan-v/copperheados-stack/releases) page.

## Prerequisites
You'll need AWS CLI credentials setup with 'AdministratorAccess': https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html

## Deployment
* Deploy environment for Pixel XL in AWS (marlin)

    ```sh
    ./copperheados-stack --region us-west-2 --name copperheados-dan --device marlin
    ```

* Deploy environment for Pixel in AWS (sailfish)

    ```sh
    ./copperheados-stack --region us-west-2 --name copperheados-dan --device sailfish
    ```

* Deploy environment for Pixel 2 XL in AWS (taimen)

    ```sh
    ./copperheados-stack --region us-west-2 --name copperheados-dan --device taimen
    ```

* Deploy environment for Pixel 2 in AWS (walleye)

    ```sh
    ./copperheados-stack --region us-west-2 --name copperheados-dan --device walleye
    ```

* Remove environment and all AWS resources

    ```sh
    ./copperheados-stack --remove --region us-west-2 --name copperheados-dan
    ```

## First Time Setup After Deployment
* Initial build should automatically kick off (it will take a few hours).
* After build finishes, a factory image should be uploaded to the S3 bucket '\<stackname>-release'. From this bucket, download the file '\<device>-factory-latest.tar.xz'. 
* Use this factory image and follow the instructions on flashing your device: https://copperhead.co/android/docs/install
* After successfully flashing your device, you will now be running CopperheadOS and all future updates will happen through built in OTA mechanism.

## Updating to a New Version
* Just download the new version and run the same command used previously (e.g. ./copperheados-stack --region us-west-2 --name copperheados-dan --device marlin) to apply the updates

## FAQ
1. <b>Should I use copperheados-stack?</b> That's up to you. Use at your own risk.

## Powered by
* All credit goes to Huimin Zhang. He is the original author of the underlying build script!
* [Terraform](https://www.terraform.io/) 

## Build From Source

  ```sh
  make tools && make
  ```

## To Do
* Restrict created IAM roles to minimum required privileges (currently all admin)