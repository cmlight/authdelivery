

start with the full bid request as received

retrieve all schain nodes

For each node, 
    parse the individual node into a temp data structure

    add the protected fields and replacement key/value pairs to the node snapshot


for each protected field, fetch the value from the full bid request and store in a map

for each parsed node
    create the string for the node's own protected asi/sid contents

    for each protected field,
        search forward to find the next replacement value, if any

        if found, use it
        else, use the value found in the bid request