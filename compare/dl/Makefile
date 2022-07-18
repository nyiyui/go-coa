all: gos_ gop_ pys_ pyp_ coas_ coap_;

gos_:
	go build -o go_exec go/main.go
	time ./go_exec dl_list.txt
	rm go_exec

gop_:
	go build -o gop_exec gop/main.go
	time ./gop_exec dl_list.txt
	rm gop_exec

pys_:
	time python3 py dl_list.txt

pyp_:
	time python3 pyp dl_list.txt

coas_:
	go build -o coa_exec ../../try2/cmd/main.go
	time ./coa_exec --parallel=false --path=coa/main.coa dl_list.txt
	rm coa_exec

coap_:
	go build -o coa_exec ../../try2/cmd/main.go
	time ./coa_exec --parallel=true --path=coa/main.coa dl_list.txt
	rm coa_exec
