
#!/bin/bash

read -p "number of nodes: " number_of_nodes

#creates the address-book.txt if it does not exists
touch address-book.txt
mkdir -p output

for (( i=1; i<=$number_of_nodes; i++ ))
do  
   ./server < address-book.txt >> address-book.txt 2> output/"$i.log" &
done