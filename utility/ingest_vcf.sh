#!/bin/sh
# NAME
#     ingest_vcf.sh - download, filter, and subsample G. max VCFs
# 
# SYNOPSIS
#     ingest_vcf.sh VCF_URL_SUFFIX [-qual_min QUAL]
#
# STDOUT
#     VCF
#
# EXAMPLE
#     ingest_vcf.sh Wm82.gnm2.div.Valliyodan_Brown_2021/glyma.Wm82.gnm2.div.Valliyodan_Brown_2021.USB481.vcf.gz -qual_min 1000 > glyma.Wm82.gnm2.div.Valliyodan_Brown_2021.USB481_sub25k.vcf

set -o errexit -o nounset -o pipefail

readonly DATA_STORE_URL=https://www.soybase.org/data/v2/Glycine/max/diversity/

readonly vcf=${1}
shift

wget -O - ${DATA_STORE_URL}/${vcf} |
  gzip -dc |
    # load only chromosomes
    awk 'tolower($1) !~ /scaff|contig|nc_/' |
      subsample_vcf.pl ${@:-}
