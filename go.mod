module github.com/diamnet/go

go 1.16

require (
	cloud.google.com/go/firestore v1.5.0 // indirect
	firebase.google.com/go v3.12.0+incompatible
	github.com/BurntSushi/toml v0.3.1
	github.com/Masterminds/squirrel v1.5.0
	github.com/Microsoft/go-winio v0.4.14
	github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/ajg/form v0.0.0-20160822230020-523a5da1a92f // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/aws/aws-sdk-go v1.39.5
	github.com/elazarl/go-bindata-assetfs v1.0.0
	github.com/fatih/structs v1.0.0 // indirect
	github.com/gavv/monotime v0.0.0-20161010190848-47d58efa6955 // indirect
	github.com/getsentry/raven-go v0.0.0-20160805001729-c9d3cc542ad1
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/go-errors/errors v0.0.0-20150906023321-a41850380601
	github.com/gobuffalo/packr v1.12.1 // indirect
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/google/go-querystring v0.0.0-20160401233042-9235644dd9e5 // indirect
	github.com/google/uuid v1.2.0
	github.com/gorilla/schema v1.1.0
	github.com/graph-gophers/graphql-go v0.0.0-20190225005345-3e8838d4614c
	github.com/guregu/null v2.1.3-0.20151024101046-79c5bd36b615+incompatible
	github.com/holiman/uint256 v1.2.0
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/imkira/go-interpol v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jarcoal/httpmock v0.0.0-20161210151336-4442edb3db31
	github.com/jmoiron/sqlx v1.2.0
	github.com/klauspost/compress v0.0.0-20161106143436-e3b7981a12dd // indirect
	github.com/klauspost/cpuid v0.0.0-20160302075316-09cded8978dc // indirect
	github.com/klauspost/crc32 v0.0.0-20161016154125-cb6bfca970f6 // indirect
	github.com/lib/pq v1.2.0
	github.com/magiconair/properties v1.5.4 // indirect
	github.com/manucorporat/sse v0.0.0-20160126180136-ee05b128a739
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v0.0.0-20150613213606-2caf8efc9366 // indirect
	github.com/moul/http2curl v0.0.0-20161031194548-4e24498b31db // indirect
	github.com/onsi/ginkgo v1.7.0
	github.com/onsi/gomega v1.4.3
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pelletier/go-toml v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4
	github.com/rs/cors v0.0.0-20160617231935-a62a804a8a00
	github.com/rs/xhandler v0.0.0-20160618193221-ed27b6fd6521 // indirect
	github.com/rubenv/sql-migrate v0.0.0-20190717103323-87ce952f7079
	github.com/segmentio/go-loggly v0.5.1-0.20171222203950-eb91657e62b2
	github.com/sergi/go-diff v0.0.0-20161205080420-83532ca1c1ca // indirect
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/sirupsen/logrus v1.4.1
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337 // indirect
	github.com/spf13/cast v0.0.0-20150508191742-4d07383ffe94 // indirect
	github.com/spf13/cobra v0.0.0-20160830174925-9c28e4bbd74e
	github.com/spf13/jwalterweatherman v0.0.0-20141219030609-3d60171a6431 // indirect
	github.com/spf13/pflag v0.0.0-20161005214240-4bd69631f475
	github.com/spf13/viper v0.0.0-20150621231900-db7ff930a189
	github.com/diamnet/go-xdr v0.0.0-20211103144802-8017fc4bdfee
	github.com/diamnet/throttled v2.2.3-0.20190823235211-89d75816f59d+incompatible
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tyler-smith/go-bip39 v0.0.0-20180618194314-52158e4697b8
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v0.0.0-20170109085056-0a7f0a797cd6 // indirect
	github.com/xdrpp/goxdr v0.1.1
	github.com/xeipuuv/gojsonpointer v0.0.0-20151027082146-e0fe6f683076 // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20150808065054-e02fc20de94c // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20161231055540-f06f290571ce // indirect
	github.com/yalp/jsonpath v0.0.0-20150812003900-31a79c7593bb // indirect
	github.com/yudai/gojsondiff v0.0.0-20170107030110-7b1b7adf999d // indirect
	github.com/yudai/golcs v0.0.0-20150405163532-d1c525dea8ce // indirect
	github.com/yudai/pp v2.0.1+incompatible // indirect
	github.com/ziutek/mymysql v1.5.4 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	google.golang.org/api v0.50.0
	gopkg.in/gavv/httpexpect.v1 v1.0.0-20170111145843-40724cf1e4a0
	gopkg.in/gorp.v1 v1.7.1 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1
	gopkg.in/tylerb/graceful.v1 v1.2.13
)
