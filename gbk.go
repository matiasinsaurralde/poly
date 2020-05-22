package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

//used in parseLocus function though it could be useful elsewhere.
var genbankDivisions = []string{
	"PRI", //primate sequences
	"ROD", //rodent sequences
	"MAM", //other mamallian sequences
	"VRT", //other vertebrate sequences
	"INV", //invertebrate sequences
	"PLN", //plant, fungal, and algal sequences
	"BCT", //bacterial sequences
	"VRL", //viral sequences
	"PHG", //bacteriophage sequences
	"SYN", //synthetic sequences
	"UNA", //unannotated sequences
	"EST", //EST sequences (expressed sequence tags)
	"PAT", //patent sequences
	"STS", //STS sequences (sequence tagged sites)
	"GSS", //GSS sequences (genome survey sequences)
	"HTG", //HTG sequences (high-throughput genomic sequences)
	"HTC", //unfinished high-throughput cDNA sequencing
	"ENV", //environmental sampling sequences
}

//used in feature check functions.
var genbankTopLevelFeatures = []string{
	"LOCUS",
	"DEFINITION",
	"ACCESSION",
	"VERSION",
	"KEYWORDS",
	"SOURCE",
	"REFERENCE",
	"FEATURES",
	"ORIGIN",
}

//used in feature check functions.
var genbankSubLevelFeatures = []string{
	"ORGANISM",
	"AUTHORS",
	"TITLE",
	"JOURNAL",
	"PUBMED",
	"REMARK",
}

//all gene feature types in genbank
var genbankGeneFeatureTypes = []string{
	"assembly_gap",
	"C_region",
	"CDS",
	"centromere",
	"D-loop",
	"D_segment",
	"exon",
	"gap",
	"gene",
	"iDNA",
	"intron",
	"J_segment",
	"mat_peptide",
	"misc_binding",
	"misc_difference",
	"misc_feature",
	"misc_recomb",
	"misc_RNA",
	"misc_structure",
	"mobile_element",
	"modified_base",
	"mRNA",
	"ncRNA",
	"N_region",
	"old_sequence",
	"operon",
	"oriT",
	"polyA_site",
	"precursor_RNA",
	"prim_transcript",
	"primer_bind",
	"propeptide",
	"protein_bind",
	"regulatory",
	"repeat_region",
	"rep_origin",
	"rRNA",
	"S_region",
	"sig_peptide",
	"source",
	"stem_loop",
	"STS",
	"telomere",
	"tmRNA",
	"transit_peptide",
	"tRNA",
	"unsure",
	"V_region",
	"V_segment",
	"variation",
	"3'UTR",
	"5'UTR",
}

// all genbank feature qualifiers
var genbankGeneQualifierTypes = []string{
	"/allele=",
	"/altitude=",
	"/anticodon=",
	"/artificial_location",
	"/bio_material=",
	"/bound_moiety=",
	"/cell_line=",
	"/cell_type=",
	"/chromosome=",
	"/citation=",
	"/clone=",
	"/clone_lib=",
	"/codon_start=",
	"/collected_by=",
	"/collection_date=",
	"/compare=",
	"/country=",
	"/cultivar=",
	"/culture_collection=",
	"/db_xref=",
	"/dev_stage=",
	"/direction=",
	"/EC_number=",
	"/ecotype=",
	"/environmental_sample",
	"/estimated_length=",
	"/exception=",
	"/experiment=",
	"/focus",
	"/frequency=",
	"/function=",
	"/gap_type=",
	"/gene=",
	"/gene_synonym=",
	"/germline",
	"/haplogroup=",
	"/haplotype=",
	"/host=",
	"/identified_by=",
	"/inference=",
	"/isolate=",
	"/isolation_source=",
	"/lab_host=",
	"/lat_lon=",
	"/linkage_evidence=",
	"/locus_tag=",
	"/macronuclear",
	"/map=",
	"/mating_type=",
	"/metagenome_source",
	"/mobile_element_type=",
	"/mod_base=",
	"/mol_type=",
	"/ncRNA_class=",
	"/note=",
	"/number=",
	"/old_locus_tag=",
	"/operon=",
	"/organelle=",
	"/organism=",
	"/partial",
	"/PCR_conditions=",
	"/PCR_primers=",
	"/phenotype=",
	"/plasmid=",
	"/pop_variant=",
	"/product=",
	"/protein_id=",
	"/proviral",
	"/pseudo",
	"/pseudogene=",
	"/rearranged",
	"/replace=",
	"/ribosomal_slippage",
	"/rpt_family=",
	"/rpt_type=",
	"/rpt_unit_range=",
	"/rpt_unit_seq=",
	"/satellite=",
	"/segment=",
	"/serotype=",
	"/serovar=",
	"/sex=",
	"/specimen_voucher=",
	"/standard_name=",
	"/strain=",
	"/sub_clone=",
	"/submitter_seqid=",
	"/sub_species=",
	"/sub_strain=",
	"/tag_peptide=",
	"/tissue_lib=",
	"/tissue_type=",
	"/transgenic",
	"/translation=",
	"/transl_except=",
	"/transl_table=",
	"/trans_splicing",
	"/type_material=",
	"/variety=",
}

// indeces for random points of interests on a gbk line.
const metaIndex = 0
const subMetaIndex = 5
const qualifierIndex = 21

func quickMetaCheck(line string) bool {
	flag := false
	if string(line[metaIndex]) != " " {
		flag = true
	}
	return flag
}

func quickSubMetaCheck(line string) bool {
	flag := false

	if string(line[metaIndex]) == " " && string(line[subMetaIndex]) != " " {
		flag = true
	}
	return flag
}

func quickFeatureCheck(line string) bool {
	flag := false

	if string(line[metaIndex]) == " " && string(line[subMetaIndex]) != " " {
		flag = true
	}
	return flag
}

func quickQualifierCheck(line string) bool {
	flag := false

	if string(line[metaIndex]) == " " && string(line[subMetaIndex]) == " " && string(line[qualifierIndex]) == "/" {
		flag = true
	}
	return flag

}

func quickQualifierSubLineCheck(line string) bool {
	flag := false

	if string(line[metaIndex]) == " " && string(line[subMetaIndex]) == " " && string(line[qualifierIndex]) != "/" && string(line[qualifierIndex-1]) == " " {
		flag = true
	}
	return flag
}

// checks for only top level features in genbankTopLevelFeatures array
func topLevelFeatureCheck(featureString string) bool {
	flag := false
	cleanedFeatureString := strings.TrimSpace(featureString)
	for _, feature := range genbankTopLevelFeatures {
		if feature == cleanedFeatureString {
			flag = true
			break
		}
	}
	return flag
}

// checks for only sub level features in genbankSubLevelFeatures array
func subLevelFeatureCheck(featureString string) bool {
	flag := false
	cleanedFeatureString := strings.TrimSpace(featureString)
	for _, feature := range genbankSubLevelFeatures {
		if feature == cleanedFeatureString {
			flag = true
			break
		}
	}
	return flag
}

// checks for both sub and top level features in genbankSubLevelFeatures and genbankTopLevelFeatures array
func allLevelFeatureCheck(featureString string) bool {
	flag := false
	cleanedFeatureString := strings.TrimSpace(featureString)
	if subLevelFeatureCheck(cleanedFeatureString) || topLevelFeatureCheck(cleanedFeatureString) {
		flag = true
	}
	return flag
}

// will eventually refactor all checks into one function.
func geneFeatureTypeCheck(featureString string) bool {
	flag := false
	cleanedFeatureString := strings.TrimSpace(featureString)
	for _, feature := range genbankGeneFeatureTypes {
		if feature == cleanedFeatureString {
			flag = true
			break
		}
	}
	return flag
}

func geneQualifierTypeCheck(featureString string) bool {
	flag := false
	cleanedFeatureString := strings.TrimSpace(strings.SplitAfter(featureString, "=")[0])
	for _, feature := range genbankGeneQualifierTypes {
		if feature == cleanedFeatureString {
			flag = true
			break
		}
	}
	return flag
}

func allGeneTypeCheck(featureString string) bool {
	flag := false
	cleanedFeatureString := strings.TrimSpace(featureString)
	if geneQualifierTypeCheck(cleanedFeatureString) || topLevelFeatureCheck(cleanedFeatureString) {
		flag = true
	}
	return flag
}

// parses locus from provided string.
func parseLocus(locusString string) Locus {
	locus := Locus{}
	locusSplit := strings.Split(strings.TrimSpace(locusString), " ")
	var filteredLocusSplit []string
	for i := range locusSplit {
		if locusSplit[i] != "" {
			filteredLocusSplit = append(filteredLocusSplit, locusSplit[i])
		}
	}
	locus.Name = filteredLocusSplit[1]
	locus.SequenceLength = strings.Join([]string{filteredLocusSplit[2], filteredLocusSplit[3]}, " ")
	locus.MoleculeType = filteredLocusSplit[4]
	if filteredLocusSplit[5] == "circular" || filteredLocusSplit[5] == "linear" {
		if filteredLocusSplit[5] == "circular" {
			locus.Circular = true
		} else {
			locus.Circular = false
		}
		locus.GenBankDivision = filteredLocusSplit[6]
		locus.ModDate = filteredLocusSplit[7]
	} else {
		locus.Circular = false
		locus.GenBankDivision = filteredLocusSplit[5]
		locus.ModDate = filteredLocusSplit[6]
	}
	return locus
}

// really important helper function. It finds sublines of a feature and joins them.
func joinSubLines(splitLine, subLines []string) string {
	base := strings.TrimSpace(strings.Join(splitLine[1:], " "))

	for _, subLine := range subLines {
		featureSplitLine := strings.Split(strings.TrimSpace(subLine), " ")
		headString := featureSplitLine[0]
		if !allLevelFeatureCheck(headString) {
			base = strings.TrimSpace(strings.TrimSpace(base) + " " + strings.TrimSpace(subLine))
		} else {
			break
		}
	}
	return base
}

// get organism name and source. Doesn't use joinSubLines.
func getSourceOrganism(splitLine, subLines []string) (string, string) {
	source := strings.TrimSpace(strings.Join(splitLine[1:], " "))
	var organism string
	for numSubLine, subLine := range subLines {
		headString := strings.Split(strings.TrimSpace(subLine), " ")[0]
		if string(subLine[0]) == " " && headString != "ORGANISM" {
			source = strings.TrimSpace(strings.TrimSpace(source) + " " + strings.TrimSpace(subLine))
		} else {
			organismSubLines := subLines[numSubLine+1:]
			organismSplitLine := strings.Split(strings.TrimSpace(subLine), " ")
			organism = joinSubLines(organismSplitLine, organismSubLines)
			break
		}
	}
	return source, organism
}

// gets a single reference. Parses headstring and the joins sub lines based on feature.
func getReference(splitLine, subLines []string) Reference {
	base := strings.TrimSpace(strings.Join(splitLine[1:], " "))
	reference := Reference{}
	reference.Index = strings.Split(base, " ")[0]
	start, _ := strconv.Atoi(strings.TrimSpace(strings.Split(base, " ")[3]))
	reference.Start = start
	end := strings.TrimSpace(strings.Split(base, " ")[5])
	end = end[0 : len(end)-1]
	reference.End, _ = strconv.Atoi(end)

	for numSubLine, subLine := range subLines {
		featureSubLines := subLines[numSubLine+1:]
		featureSplitLine := strings.Split(strings.TrimSpace(subLine), " ")
		headString := featureSplitLine[0]
		if topLevelFeatureCheck(headString) {
			break
		}
		switch headString {
		case "AUTHORS":
			reference.Authors = joinSubLines(featureSplitLine, featureSubLines)
		case "TITLE":
			reference.Title = joinSubLines(featureSplitLine, featureSubLines)
		case "JOURNAL":
			reference.Journal = joinSubLines(featureSplitLine, featureSubLines)
		case "PUBMED":
			reference.PubMed = joinSubLines(featureSplitLine, featureSubLines)
		case "REMARK":
			reference.Remark = joinSubLines(featureSplitLine, featureSubLines)
		default:
			break
		}

	}
	return reference
}

func getFeatures(lines []string) []Feature {
	lineIndex := 0
	features := []Feature{}

	// regex to remove quotes and slashes from qualifiers
	reg, _ := regexp.Compile("[\"/\n]+")

	// go through every line.
	for lineIndex < len(lines) {
		line := lines[lineIndex]
		// This is a break to ensure that cursor doesn't go beyond ORIGIN which is the last top level feature.
		// This could pick up random sequence strings that aren't helpful and will mess with parser.
		// DO NOT MOVE/REMOVE WITHOUT CAUSE AND CONSIDERATION
		if quickMetaCheck(line) || !quickFeatureCheck(line) {
			break
		}

		feature := Feature{}

		// split the current line for feature type and location fields.
		splitLine := strings.Split(strings.TrimSpace(line), " ")

		// assign type and location to feature.
		feature.Type = strings.TrimSpace(splitLine[0])
		feature.Location = strings.TrimSpace(splitLine[len(splitLine)-1])

		// initialize attributes.
		feature.Attributes = make(map[string]string)

		// end of feature declaration line. Bump to next line and begin looking for qualifiers.
		lineIndex++
		line = lines[lineIndex]

		// loop through potential qualifiers. Break if not a qualifier or sub line.
		// Definition of qualifiers here: http://www.insdc.org/files/feature_table.html#3.3
		for {
			// make sure what we're parsing is a qualifier. Break if not.
			// keeping out of normal if else pattern because of phantom brackets that are hard to trace.
			if !quickQualifierCheck(line) {
				break
			}
			qualifier := line

			// end of qualifier declaration line. Bump to next line and begin looking for qualifier sublines.
			lineIndex++
			line = lines[lineIndex]

			// loop through any potential continuing lines of qualifiers. Break if not.
			for {
				// keeping out of normal if else pattern because of phantom brackets that are hard to trace.
				if !quickQualifierSubLineCheck(line) {
					break
				}
				//append to current qualifier
				qualifier += strings.TrimSpace(line)

				// nextline
				lineIndex++
				line = lines[lineIndex]
			}
			//add qualifier to feature.
			attributeSplit := strings.Split(reg.ReplaceAllString(qualifier, ""), "=")
			attributeLabel := strings.TrimSpace(attributeSplit[0])
			var attributeValue string
			if len(attributeSplit) < 2 {
				attributeValue = ""
			} else {
				attributeValue = strings.TrimSpace(attributeSplit[1])
			}
			feature.Attributes[attributeLabel] = attributeValue
		}

		//append the parsed feature to the features list to be returned.
		features = append(features, feature)

	}
	return features
}

// takes every line after origin feature and removes anything that isn't in the alphabet. Returns sequence.
func getSequence(subLines []string) Sequence {
	sequence := Sequence{}
	var sequenceBuffer bytes.Buffer
	reg, err := regexp.Compile("[^a-zA-Z]+")
	if err != nil {
		log.Fatal(err)
	}
	for _, subLine := range subLines {
		sequenceBuffer.WriteString(subLine)
	}
	sequence.Sequence = reg.ReplaceAllString(sequenceBuffer.String(), "")
	return sequence
}

func parseGbk(path string) AnnotatedSequence {

	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}

	bio := bufio.NewReader(f)

	var lines []string
	i := 0

	// Read all lines of the file into buffer
	for {
		line, err := bio.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		sline := strings.TrimRight(string(line), `\n`)
		lines = append(lines, sline)
		i++
	}
	// End read of file into buffer

	// Create meta struct
	meta := Meta{}

	// Create features struct
	features := []Feature{}

	// Create sequence struct
	sequence := Sequence{}

	for numLine := 0; numLine < len(lines); numLine++ {
		line := lines[numLine]
		splitLine := strings.Split(line, " ")
		subLines := lines[numLine+1:]

		// This is to keep the cursor from scrolling to the bottom another time after getSequence() is called.
		// Break has to be in scope and can't be called within switch statement.
		// Otherwise it will just break the switch which is redundant.
		sequenceBreakFlag := false
		if sequenceBreakFlag == true {
			break
		}

		switch splitLine[0] {

		case "":
			continue
		case "LOCUS":
			meta.Locus = parseLocus(line)
		case "DEFINITION":
			meta.Definition = joinSubLines(splitLine, subLines)
		case "ACCESSION":
			meta.Accession = joinSubLines(splitLine, subLines)
		case "VERSION":
			meta.Version = joinSubLines(splitLine, subLines)
		case "KEYWORDS":
			meta.Keywords = joinSubLines(splitLine, subLines)
		case "SOURCE":
			meta.Source, meta.Organism = getSourceOrganism(splitLine, subLines)
		case "REFERENCE":
			meta.References = append(meta.References, getReference(splitLine, subLines))
			continue
		case "FEATURES":
			features = getFeatures(subLines)
		case "ORIGIN":
			sequence = getSequence(subLines)
			sequenceBreakFlag = true
		default:
			continue
		}

	}
	var annotatedSequence AnnotatedSequence
	annotatedSequence.Meta = meta
	annotatedSequence.Features = features
	annotatedSequence.Sequence = sequence

	return annotatedSequence
}
