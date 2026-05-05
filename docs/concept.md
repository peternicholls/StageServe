# StageServe: Concepts

## Overview

StageServe is a local development tool that helps you run your projects in a local environment that mimics production. It provides an easy-to-use terminal interface to set up and manage your projects, allowing you to focus on development without worrying about the underlying infrastructure.

At the heart of StageServe is **easy mode**: a guided terminal interface for people who want a local project running without learning server terminology first. A user should be able to open a terminal in a project folder, run `stage`, review the proposed project name, web folder, and local address, and confirm the setup. StageServe then prepares the local machine, creates the project settings file, adds the configured local hostname suffix such as `.develop` to local DNS, and runs the project behind a StageServe-managed local server.

Easy mode is not a simplified command reference. It is a step-by-step flow that tells the user what is true, what StageServe recommends next, what will change, and how to undo or troubleshoot when something does not work. Advanced users still keep direct command access and custom settings, but the first screen must be clear enough for someone who does not know Docker, DNS, gateways, containers, or state files.
