grpc-gen:
	protoc --plugin=protoc-gen-echo=../protoc-gen-echo --proto_path=../third_party --proto_path=. --go_out=. --echo_out=.  helloworld.proto


	protoc  --proto_path=./third_party --proto_path=.  --go_out=. --go_opt=paths=source_relative  --customtag_out=.  custom_options.proto helloworld.proto

	protoc --plugin=protoc-gen-echo=../protoc-gen-echo/protoc-gen-echo --proto_path=./third_party --proto_path=. --go_out=. --go_opt=paths=source_relative --echo_out=.  --customtag_out=customtag:../customtag/customtag helloworld.proto
	
	protoc --plugin=protoc-gen-customtag=../customtag/protoc-gen-customtag --plugin=protoc-gen-echo=../protoc-gen-echo/protoc-gen-echo --proto_path=./third_party --proto_path=. --go_out=. --go_opt=paths=source_relative --echo_out=.  --customtag_out=. helloworld.proto

	protoc --plugin=protoc-gen-customtag=../customtag/customtag --plugin=protoc-gen-echo=../protoc-gen-echo/protoc-gen-echo --proto_path=./third_party --proto_path=. --go_out=. --go_opt=paths=source_relative --echo_out=.  --customtag_out=customtag:. helloworld.proto

	protoc --go_out=. --plugin=protoc-gen-customtag=/path/to/your/protoc-gen-customtag --customtag_out=output_directory your_proto_file.proto

	protoc --go_out=. --customtag_out=customtag:output_directory your_proto_file.proto