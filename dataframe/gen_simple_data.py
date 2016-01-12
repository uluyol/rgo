import sys

outFile = sys.argv[1]
types = sys.argv[2:]

with open(outFile, "w") as w:
	w.write("package dataframe\n\n")
	w.write('import "fmt"\n\n')
	w.write("func ensureSimpleData(x SimpleData) {\n")
	w.write("\tswitch x.(type) {\n")
	for t in types:
		w.write("\tcase %s: // no error\n" % (t,))
	w.write("\tdefault:\n")
	w.write('\t\tpanic(fmt.Sprintf("%s is not a valid SimpleData value", x))\n')
	w.write("\t}\n}\n\n")

	w.write("func slicePtrOf(dtype string) (interface{}, error) {\n")
	w.write("\tswitch dtype {\n")
	w.write('\tcase "empty":\n')
	w.write("\t\treturn new([]SimpleData), nil\n")
	for t in types:
		w.write('\tcase "%s":\n' % (t,))
		w.write("\t\treturn new([]%s), nil\n" % (t,))
	w.write("\t}\n")
	w.write('\treturn nil, fmt.Errorf("invalid data type %q for SimpleData", dtype)\n')
	w.write("}\n")
