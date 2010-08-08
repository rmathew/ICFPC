JAVA_HOME=/usr/local/jdk
JAVA=$(JAVA_HOME)/bin/java
JAVAC=$(JAVA_HOME)/bin/javac

CLASSPATH=.

CLASSES=\
	FuelFactory.class \
	Utils.class \
	Circuit.class \
	Lilo.class \
	Stitch.class \
	Cir2Dot.class \

INPUT=server.inp
TARGET=key.tgt
INP2TGT_CIR=inp2tgt.cir
CIRCUIT=circuits/L2L_1.cir

.PHONY: all run stitch todot clean

%.class: %.java
	$(JAVAC) -cp $(CLASSPATH) $<

all: $(CLASSES)

run: $(CLASSES)
	@echo =====
	@cat $(CIRCUIT)
	@echo =====
	$(JAVA) -cp $(CLASSPATH) FuelFactory $(INPUT) $(CIRCUIT)

stitch: $(CLASSES)
	$(JAVA) -cp $(CLASSPATH) Stitch $(INPUT) $(TARGET) $(INP2TGT_CIR)

todot: $(CLASSES)
	$(JAVA) -cp $(CLASSPATH) Cir2Dot $(CIRCUIT)

clean:
	rm -f $(CLASSES)
