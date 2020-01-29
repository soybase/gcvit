/*
	route handlers for the gcvit server
*/

package gcvit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/awilkey/bio-format-tools-go/gff"
	"github.com/awilkey/bio-format-tools-go/vcf"
	"github.com/valyala/fasthttp"
	"math"
	"strconv"
	"strings"
	"time"
)

// experiment that stores all the allowed experiments
var experiments map[string]DataFiles

//init does initialization, only runs first time gcvit is called
func init() {
	experiments = make(map[string]DataFiles)
}

// GetExperiments is a GET path that returns a JSON object that represents all the currently loaded datasets GT field headers
func GetExperiments(ctx *fasthttp.RequestCtx) {
	//start time for logging
	start := time.Now()
	// Populate experiments if it hasn't been already
	if len(experiments) == 0 {
		PopulateExperiments()
	}

	//Iterate through experiments and build response
	opts := make([]ExpData, len(experiments))
	i := 0
	for key := range experiments {
		exp := ExpData{Value: key, Label: experiments[key].Name}
		opts[i] = exp
		i++
	}

	//Response
	optsJson, _ := json.Marshal(opts)
	ctx.SetContentType("application/json; charset=utf8")
	fmt.Fprintf(ctx, "%s", optsJson)
	//Log response time
	ctx.Logger().Printf("%dns", time.Now().Sub(start).Nanoseconds())
}

// GetExperiments is a GET path that returns a JSON representation of all the passed experiment's
func GetExperiment(ctx *fasthttp.RequestCtx) {
	//Start time for logging
	start := time.Now()
	//Parse arguement from path (/api/experiment/:exp
	exp := ctx.UserValue("exp").(string)
	//Populate experiments if it hasn't been populated already
	if len(experiments) == 0 {
		PopulateExperiments()
	}

	//Iterate through passed experiment and build response of GT headers
	gt := make([]ExpData, len(experiments[exp].Genotypes))
	for i, v := range experiments[exp].Genotypes {
		gt[i] = ExpData{Value: v, Label: v}
	}

	//Response
	gtJson, _ := json.Marshal(gt)
	ctx.SetContentType("application/json; charset=utf8")
	fmt.Fprintf(ctx, "%s", gtJson)
	//Log response time
	ctx.Logger().Printf("%dns", time.Now().Sub(start).Nanoseconds())
}

//GenerateGFF takes a post request for a given vcf file and returns a GFF
func GenerateGFF(ctx *fasthttp.RequestCtx) {
	//Log request received
	ctx.Logger().Printf("Begin request for: %s", ctx.PostArgs())
	start := time.Now()
	//Struct for holding Post Request
	req := &struct {
		Ref     string
		Variant []string
		Bin     int
	}{}

	//parse reference, both ref and variant are in the form "<exp>:<gt>"
	//Peek is used here, as current GCViT only supports a single reference
	req.Ref = string(ctx.PostArgs().Peek("Ref"))

	//parse variant(s)
	vnts := ctx.PostArgs().PeekMulti("Variant")
	for _, v := range vnts {
		req.Variant = append(req.Variant, string(v))
	}

	//parse bin size if available, if not passed, default to 500000 bases
	if bSize, _ := strconv.Atoi(string(ctx.PostArgs().Peek("Bin"))); bSize > 0 {
		req.Bin = bSize
	} else {
		req.Bin = 500000
	}

	ref := strings.Split(req.Ref, ":")
	vnt := make(map[string][]string, len(req.Variant))
	vntOrder := make(map[int][]string, len(req.Variant))
	for i := range req.Variant {
		vt := strings.Split(req.Variant[i], ":")
		if _, ok := vnt[vt[0]]; !ok {
			vnt[vt[0]] = []string{vt[1]}
		} else {
			vnt[vt[0]] = append(vnt[vt[0]], vt[1])
		}
		vntOrder[i] = []string{vt[0], vt[1]}
	}

	r, err := ReadFile(experiments[ref[0]].Location, experiments[ref[0]].Gzip)
	if err != nil {
		panic(fmt.Errorf("problem reading reference genotype's file: %s \n", err))
	}

	var b bytes.Buffer
	writer, err := gff.NewWriter(&b)

	if err != nil {
		panic(fmt.Errorf("problem opening gff writer: %s \n", err))
	}

	ctg := make(map[string]int)
	for i := range r.Header.Contigs {
		ctgLen, _ := strconv.Atoi(r.Header.Contigs[i].Optional["length"])
		ctg[r.Header.Contigs[i].Id] = ctgLen
	}

	sameCtr := make(map[string]int, len(vnt[ref[0]])+1)
	diffCtr := make(map[string]int, len(vnt[ref[0]])+1)
	totalCtr := make(map[string]int, len(vnt[ref[0]])+3)
	totalCtr[ref[1]] = 0
	totalCtr["undefined"] = 0
	totalCtr["value"] = 0
	sameCtr["value"] = 0
	diffCtr["value"] = 0

	for i := range vnt[ref[0]] {
		gt := vnt[ref[0]][i]
		sameCtr[gt] = 0
		diffCtr[gt] = 0
		totalCtr[gt] = 0
	}

	var feat *vcf.Feature
	var readErr error
	var contig string
	var stepSize int
	if req.Bin > 0 {
		stepSize = req.Bin
	} else {
		stepSize = 500000
	}
	stepCt := 0
	stepVal := 0

	for readErr == nil {
		feat, readErr = r.Read()
		if feat != nil {
			gt, _ := feat.SingleGenotype(ref[1], r.Header.Genotypes)
			rt, _ := feat.MultipleGenotypes(vnt[ref[0]], r.Header.Genotypes)
			//reset contig based features, assuming that file is sorted by contig and ascending position
			// when contig changes or you step outside of current bin
			if feat.Pos > uint64(stepCt*stepVal) || contig != feat.Chrom {
				if stepCt > 0 {
					end := uint64(stepCt) * uint64(stepVal)
					if ctg[contig] > 0 && end > uint64(ctg[contig]) {
						end = uint64(ctg[contig])
					}
					gffLine := gff.Feature{
						Seqid:      contig,
						Source:     "soybase",
						Type:       "same",
						Start:      uint64((stepCt-1)*stepVal + 1),
						End:        end,
						Score:      gff.MissingScoreField,
						Strand:     "+",
						Phase:      gff.MissingPhaseField,
						Attributes: map[string]string{"ID": fmt.Sprintf("%s.%d", "same", stepCt)},
					}
					for class, val := range sameCtr {
						gffLine.Attributes[class] = strconv.Itoa(val)
					}
					writer.WriteFeature(&gffLine)

					gffLine.Type = "diff"
					gffLine.Attributes["ID"] = fmt.Sprintf("%s.%d", "diff", stepCt)
					for class, val := range diffCtr {
						gffLine.Attributes[class] = strconv.Itoa(val)
					}
					writer.WriteFeature(&gffLine)

					gffLine.Type = "total"
					gffLine.Attributes["ID"] = fmt.Sprintf("%s.%d", "total", stepCt)
					for class, val := range totalCtr {
						gffLine.Attributes[class] = strconv.Itoa(val)
					}
					writer.WriteFeature(&gffLine)

					//Reset counters
					for val := range totalCtr {
						totalCtr[val] = 0
					}
					for val := range sameCtr {
						sameCtr[val] = 0
					}
					for val := range diffCtr {
						diffCtr[val] = 0
					}
					stepCt = (int(feat.Pos) / stepSize) + 1
				}

				if contig != feat.Chrom {
					contig = feat.Chrom
					if ctg[contig] > 0 {
						stepVal = int(float64(ctg[contig]) / math.Ceil(float64(ctg[contig])/float64(stepSize)))
					} else {
						stepVal = stepSize
					}
					stepCt = 1
				}
			}
			gFields := gt.Fields["GT"]
			if gFields != "./." && gFields != ".|." {
				totalCtr["value"]++
				totalCtr[ref[1]]++

				for i := range rt {
					rFields := rt[i].Fields["GT"]
					id := rt[i].Id
					if rFields == "./." || rFields == ".|." {
						totalCtr["undefined"]++
					} else if gFields == rFields {
						sameCtr[id]++
						sameCtr["value"]++
						totalCtr[id]++
					} else {
						diffCtr[id]++
						diffCtr["value"]++
						totalCtr[id]++
					}
				}
			}
		}
	}

	end := stepCt * stepVal

	if ctg[contig] > 0 && end > ctg[contig] {
		end = ctg[contig]
	}

	gffLine := gff.Feature{
		Seqid:      contig,
		Source:     "soybase",
		Type:       "same",
		Start:      uint64((stepCt-1)*stepVal + 1),
		End:        uint64(end),
		Score:      gff.MissingScoreField,
		Strand:     "+",
		Phase:      gff.MissingPhaseField,
		Attributes: map[string]string{"ID": fmt.Sprintf("%s.%d", "same", stepCt)},
	}

	for class, val := range sameCtr {
		gffLine.Attributes[class] = strconv.Itoa(val)
	}
	writer.WriteFeature(&gffLine)

	gffLine.Type = "diff"
	gffLine.Attributes["ID"] = fmt.Sprintf("%s.%d", "diff", stepCt)
	for class, val := range diffCtr {
		gffLine.Attributes[class] = strconv.Itoa(val)
	}
	writer.WriteFeature(&gffLine)

	gffLine.Type = "total"
	gffLine.Attributes["ID"] = fmt.Sprintf("%s.%d", "total", stepCt)
	for class, val := range totalCtr {
		gffLine.Attributes[class] = strconv.Itoa(val)
	}
	writer.WriteFeature(&gffLine)

	ctx.SetContentType("text/plain; charset=utf8")
	fmt.Fprintf(ctx, "%s", b.String())
	//Log completed request
	ctx.Logger().Printf("Return request for %s - %dns", ctx.PostArgs(), time.Now().Sub(start).Nanoseconds())
}
