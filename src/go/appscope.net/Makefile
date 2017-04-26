API_ENDPOINT = "https://appnetscope.appspot.com"
#API_ENDPOINT = "http://localhost:8080"
API_LIST = share cert

api: rebuild
	rm -rf ./api
	for api in $(API_LIST); do \
		mkdir -p api/$${api}/v1 ;\
		curl --show-error -o \
			./api/$${api}/v1/$${api}-api.json ${API_ENDPOINT}/_ah/api/discovery/v1/apis/$${api}/v1/rest ; \
		${GOPATH}/bin/google-api-go-generator --api_pkg_base="appscope.net/api" \
		-api_json_file=./api/$${api}/v1/$${api}-api.json -gendir=./api ; \
		perl -i -pe "s/Token string/Token []byte/g;" ./api/$${api}/v1/$${api}-gen.go ;\
		perl -i -pe "s/Cert string/Cert []byte/g;" ./api/$${api}/v1/$${api}-gen.go ;\
		perl -i -pe "s/CA string/CA []byte/g;" ./api/$${api}/v1/$${api}-gen.go ;\
		perl -i -pe "s/Key string/Key []byte/g;" ./api/$${api}/v1/$${api}-gen.go ;\
	done

rebuild:
