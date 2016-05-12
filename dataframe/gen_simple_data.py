import sys

def usage():
	print("Usage: %s other=type1,type2,... numeric=typeN,typeN+1,..." % (sys.argv[0],))
	sys.exit(1)

outFile = sys.argv[1]
try:
	kv = (s.split("=") for s in sys.argv[2:])
	types = {k: v.split(",") for k, v in kv}
except:
	usage()

if len(types.keys()) != 2:
	usage()

types["all"] = types["other"] + types["numeric"]

with open(outFile, "w") as w:
	w.write("package dataframe\n\n")
	w.write('import (\n\t"fmt"\n\n\t"github.com/pkg/errors"\n)\n\n')
	w.write("// IsNumeric checks whether x is a numeric type. Currently these\n")
	w.write("// consist only of integers and floats. Complex numbers and big\n")
	w.write("// numbers are not considered numeric.\n")
	w.write("func IsNumeric(x SimpleData) bool {\n")
	w.write("\tswitch x.(type) {\n")
	for t in types["numeric"]:
		w.write("\tcase %s: // numeric\n" % (t,))
	w.write("\tdefault:\n")
	w.write("\t\treturn false\n")
	w.write("\t}\n")
	w.write("\treturn true\n")
	w.write("}\n\n")
	w.write("func ensureSimpleData(x SimpleData) {\n")
	w.write("\tswitch x.(type) {\n")
	for t in types["all"]:
		w.write("\tcase %s: // no error\n" % (t,))
	w.write("\tdefault:\n")
	w.write('\t\tpanic(fmt.Sprintf("%s is not a valid SimpleData value", x))\n')
	w.write("\t}\n}\n\n")

	w.write("func slicePtrOf(dtype string) (interface{}, error) {\n")
	w.write("\tswitch dtype {\n")
	w.write('\tcase "empty":\n')
	w.write("\t\treturn new([]SimpleData), nil\n")
	for t in types["all"]:
		w.write('\tcase "%s":\n' % (t,))
		w.write("\t\treturn new([]%s), nil\n" % (t,))
	w.write("\t}\n")
	w.write('\treturn nil, errors.Errorf("invalid data type %q for SimpleData", dtype)\n')
	w.write("}\n")
