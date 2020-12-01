#go run bkrs/choose_main_readings.go unihan/main-pinyin-by-hanzi.tsv.txt <bkrs/extracted_bkrs_readings_ndfa.tsv.txt 1>bkrs/chosen_main_readings.tsv.txt 2>bkrs/choose_main_readings.log

#tail -n +2 ~/Documentation/chinese/frequencies/subtlex_words/moved-out/SUBTLEX_CH_131210_CE.utf8 | go run tonify.go tonify-subtlex-frequencies.go >tonified-subtlex-frequencies.tsv.txt

#cat unihan/main-pinyin-by-hanzi.tsv.txt extracted_cedict_readings.tsv.txt bkrs/chosen_main_readings.tsv.txt | go run unify-pinyin-dictionaries.go | sort --field-separator=$'\t' --key=2 >unified-pinyin-dictionary.tsv.txt

#clear && clear && rlwrap go run unified-pinyin-dictionary-simple-interactive-lookuping.go cedict/tonified-subtlex-frequencies.tsv.txt unified-pinyin-dictionary.tsv.txt



#unzip -p ~/.mozilla/firefox/*.default-*/extensions/'{dedb3663-6f13-4c6c-bf0f-5bd111cb2c79}.xpi' 'data/cedict_ts.u8' | go run {cedict2tsv,tonify}.go >cedict.tsv.txt

#clear && clear && rlwrap go run cedict.tsv-simple-interactive-lookuping.go cedict.tsv.txt
