JAVA=java
JAVAC=javac

.java.class:
	$(JAVAC) $<

.SUFFIXES: .java .class

MAIN_CLASS=Rover
OTHER_CLASSES=

ALL_CLASSES=\
  $(MAIN_CLASS).class \
  $(OTHER_CLASSES) \

.PHONY: all clean run

all: $(ALL_CLASSES)

run: $(ALL_CLASSES)
	$(JAVA) -Djava.awt.headless=true $(MAIN_CLASS)

clean:
	rm -f $(ALL_CLASSES)
