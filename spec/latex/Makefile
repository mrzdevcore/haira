# Haira Language Specification Build System
# Requires: pdflatex, bibtex (optional)

MAIN = haira-spec
LATEX = pdflatex
LATEX_FLAGS = -interaction=nonstopmode -halt-on-error

# Output directory
OUT_DIR = build

# Source files
CHAPTERS = $(wildcard chapters/*.tex)
SOURCES = $(MAIN).tex $(CHAPTERS)

.PHONY: all clean view watch help

all: $(OUT_DIR)/$(MAIN).pdf

$(OUT_DIR)/$(MAIN).pdf: $(SOURCES) | $(OUT_DIR)
	@echo "Building $(MAIN).pdf..."
	$(LATEX) $(LATEX_FLAGS) -output-directory=$(OUT_DIR) $(MAIN).tex
	$(LATEX) $(LATEX_FLAGS) -output-directory=$(OUT_DIR) $(MAIN).tex
	@echo "Build complete: $(OUT_DIR)/$(MAIN).pdf"

$(OUT_DIR):
	mkdir -p $(OUT_DIR)

# Quick build (single pass, for development)
quick: | $(OUT_DIR)
	$(LATEX) $(LATEX_FLAGS) -output-directory=$(OUT_DIR) $(MAIN).tex

# Clean build artifacts
clean:
	rm -rf $(OUT_DIR)
	rm -f *.aux *.log *.out *.toc *.lof *.lot *.fls *.fdb_latexmk

# View the PDF (macOS)
view: $(OUT_DIR)/$(MAIN).pdf
	open $(OUT_DIR)/$(MAIN).pdf

# Watch for changes and rebuild (requires fswatch on macOS)
watch:
	@echo "Watching for changes..."
	fswatch -o $(SOURCES) | xargs -n1 -I{} make quick

# Draft mode (faster, no images)
draft: | $(OUT_DIR)
	$(LATEX) $(LATEX_FLAGS) -draftmode -output-directory=$(OUT_DIR) $(MAIN).tex

# Help
help:
	@echo "Haira Specification Build System"
	@echo ""
	@echo "Usage:"
	@echo "  make          Build the PDF (full build, 2 passes)"
	@echo "  make quick    Quick build (single pass)"
	@echo "  make clean    Remove build artifacts"
	@echo "  make view     Open the PDF (macOS)"
	@echo "  make watch    Watch for changes and rebuild"
	@echo "  make draft    Fast draft build (no images)"
	@echo "  make help     Show this help"
	@echo ""
	@echo "Output: $(OUT_DIR)/$(MAIN).pdf"
