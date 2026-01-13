#!/bin/bash

ZIP_NAME="lambda_function.zip"

rm -f $ZIP_NAME

zip -r $ZIP_NAME index.js flexTemplate.js node_modules package.json

echo "Zip Done $ZIP_NAME"
