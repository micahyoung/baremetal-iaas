#!/bin/bash

cat chain.ipxe | docker run --rm -i -w /work/ipxe/src ipxe bash -c "cat > chain.ipxe; make bin/ipxe.iso EMBED=chain.ipxe > /dev/null; cat bin/ipxe.iso"  > ipxe.iso
