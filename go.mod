module github.com/ftl/panacotta

go 1.14

replace github.com/gotk3/gotk3 => ../gotk3

replace github.com/ftl/gmtry => ../gmtry

replace github.com/ftl/hamradio => ../hamradio

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ftl/gmtry v0.0.0-20200425131616-16f55bac18a0
	github.com/ftl/hamradio v0.0.0-20191125175533-3ff46af6c6f4
	github.com/ftl/rigproxy v0.0.0-20200524134605-8e6f179b3a88
	github.com/gotk3/gotk3 v0.4.0
	github.com/jpoirier/gortlsdr v2.10.0+incompatible
	github.com/mjibson/go-dsp v0.0.0-20180508042940-11479a337f12
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
)
