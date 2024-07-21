#!/bin/bash

# Checkout the main branch and create a clean branch from it
git switch master
git merge --no-commit --squash --allow-unrelated-histories dev
