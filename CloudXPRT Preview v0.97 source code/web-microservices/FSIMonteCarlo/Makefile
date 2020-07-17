
CXX=icpc
COMPILER_VERSION:="$(CXX)-$(shell $(CXX) --version | head -n1 | cut -d' ' -f4)"
CXXFLAGS=-g -O3 -fimf-precision=low -qopenmp -restrict -qopt-report=5  -fimf-domain-exclusion=31 -no-prec-div  -no-prec-sqrt \
	-DCOMPILER_VERSION=\"$(COMPILER_VERSION)\" -qopt-report-phase=vec -DMKL_ILP64 -I${MKLROOT}/include

LDFLAGS=-L${TBBROOT}/lib/intel64/gcc4.4 -ltbbmalloc -Bstatic -L${MKLROOT}/lib/intel64/ -Wl,--start-group -lmkl_intel_ilp64 -lmkl_sequential -lmkl_core -Wl,--end-group

all: MonteCarloInsideBlockingDP.avx512 MonteCarloInsideBlockingDP.arch_avx2 

MonteCarloInsideBlockingDP.avx512: MonteCarloInsideBlockingDP.cpp
	$(CXX) $^ -xCORE-AVX512 -qopt-zmm-usage=high $(LDFLAGS) $(CXXFLAGS) -o $@

MonteCarloInsideBlockingDP.arch_avx2: MonteCarloInsideBlockingDP.cpp
	$(CXX) $^ -march=core-avx2 $(LDFLAGS) $(CXXFLAGS) -o $@

clean	:
	rm -f  *.optrpt MonteCarloInsideBlockingDP.avx512 MonteCarloInsideBlockingDP.arch_avx2
