# nitrous-solo

## Vendoring

```shell
# Install govendor
go get -u github.com/kardianos/govendor

# Vendor additional dependencies
script/savedeps
```

## Test

```shell
# Install Ginkgo/Gomega
go get -u github.com/onsi/ginkgo/ginkgo
go get -u github.com/onsi/gomega

# Run tests
script/test
```

- - -
Copyright (c) 2016 Nitrous, Inc. All Rights Reserved.
