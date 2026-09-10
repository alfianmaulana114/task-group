import fs from "node:fs/promises";
import { SpreadsheetFile, Workbook } from "@oai/artifact-tool";

const outputDir = "D:/alfian/Projek Web Liburan/task-group/outputs/it_budget_ap06_20260901";
const wb = Workbook.create();
const names = ["Dasbor", "Asumsi", "Input Anggaran", "Kontrol Anggaran", "Tata Kelola & Persetujuan", "Cek & Sumber"];
const sheets = Object.fromEntries(names.map((name) => [name, wb.worksheets.add(name)]));
for (const sheet of Object.values(sheets)) sheet.showGridLines = false;

const navy = "#163A5F", blue = "#1F5A99", teal = "#0F766E", green = "#16A34A";
const lightBlue = "#DCEAF7", lightTeal = "#DDF3EF", lightGray = "#F3F6F8";
const amber = "#FFF2CC", red = "#FCE4D6", white = "#FFFFFF", gray = "#5B6573";
const thin = { preset: "all", style: "thin", color: "#D9E1E8" };
const title = (sheet, range, text, subtitle) => {
  sheet.getRange(range).merge();
  sheet.getRange(range).values = [[text]];
  sheet.getRange(range).format = { fill: navy, font: { bold: true, color: white, size: 16 }, horizontalAlignment: "left", verticalAlignment: "center" };
  sheet.getRange(range).format.rowHeight = 28;
  const row = Number(range.match(/\d+/)[0]) + 1;
  const subRange = `A${row}:P${row}`;
  sheet.getRange(subRange).merge();
  sheet.getRange(subRange).values = [[subtitle]];
  sheet.getRange(subRange).format = { fill: "#EAF0F6", font: { color: gray, italic: true, size: 9 }, horizontalAlignment: "left" };
  sheet.getRange(subRange).format.rowHeight = 18;
};
const section = (sheet, range, text) => {
  sheet.getRange(range).merge(); sheet.getRange(range).values = [[text]];
  sheet.getRange(range).format = { fill: blue, font: { bold: true, color: white }, horizontalAlignment: "left" };
};
const header = (sheet, range) => { sheet.getRange(range).format = { fill: navy, font: { bold: true, color: white }, horizontalAlignment: "center", verticalAlignment: "center", wrapText: true, borders: thin }; };
const inputStyle = (sheet, range) => { sheet.getRange(range).format = { fill: amber, font: { color: "#1F2937" }, borders: thin }; };
const formulaStyle = (sheet, range) => { sheet.getRange(range).format = { fill: lightBlue, borders: thin }; };
const money = '"Rp" #,##0;[Red]("Rp" #,##0);-';
const months = ["Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"];

// ASSUMPTIONS
{
  const s = sheets["Asumsi"];
  title(s, "A1:H1", "Anggaran TI — Parameter & Asumsi", "Sel kuning adalah input yang dapat diedit. Semua nilai menggunakan IDR kecuali disebutkan lain.");
  section(s, "A4:D4", "Parameter tata kelola anggaran");
  s.getRange("A5:B12").values = [
    ["Tahun Fiskal", 2027], ["Bulan Pelaporan (1-12)", 1], ["Versi Anggaran", "Anggaran Kerja v1.0"],
    ["Skenario", "Dasar"], ["Mata Uang", "IDR"], ["Pemilik Anggaran", "Head of IT Finance"],
    ["Ambang Tinjauan", 0.05], ["Ambang Eskalasi", 0.10]
  ];
  s.getRange("A5:A12").format = { fill: lightGray, font: { bold: true }, borders: thin };
  inputStyle(s, "B5:B12");
  s.getRange("B11:B12").format.numberFormat = "0.0%";
  s.getRange("B6").dataValidation = { rule: { type: "whole", operator: "between", formula1: 1, formula2: 12 } };
  s.getRange("B8").dataValidation = { rule: { type: "list", values: ["Dasar", "Optimistis", "Pesimistis"] } };
  s.getRange("A:A").format.columnWidth = 24; s.getRange("B:B").format.columnWidth = 32;
  section(s, "F4:Q4", "Profil pembagian / alokasi bulanan");
  s.getRange("F5:Q5").values = [months]; header(s, "F5:Q5");
  s.getRange("F6:Q6").values = [[0.07,0.07,0.08,0.08,0.08,0.08,0.09,0.09,0.08,0.09,0.09,0.10]];
  inputStyle(s, "F6:Q6"); s.getRange("F6:Q6").format.numberFormat = "0.0%";
  s.getRange("F7:P7").merge(); s.getRange("F7:P7").values = [["Kontrol phasing (harus sama dengan 100%)"]];
  s.getRange("Q7").formulas = [["=ROUND(SUM(F6:Q6),6)"]]; formulaStyle(s, "F7:Q7"); s.getRange("Q7").format.numberFormat = "0.0%";
  section(s, "A15:H15", "Konvensi desain perusahaan");
  s.getRange("A16:B20").values = [
    ["Klasifikasi biaya", "Run / Change / Transform"], ["Tampilan kapitalisasi", "CapEx / OpEx"], ["Konvensi varians", "Varians positif = pembengkakan biaya"],
    ["Irama kontrol", "Tutup buku bulanan; reforecast triwulanan"], ["Alur persetujuan", "Owner -> IT Finance -> CIO -> CFO (sesuai kebutuhan)"]
  ];
  s.getRange("A16:A20").format = { fill: lightGray, font: { bold: true }, borders: thin };
  s.getRange("B16:B20").format = { borders: thin, wrapText: true };
  s.getRange("A16:B20").format.rowHeight = 30;
  s.freezePanes.freezeRows(4);
}

// BUDGET INPUT
{
  const s = sheets["Input Anggaran"];
  title(s, "A1:V1", "Anggaran TI — Rencana Tahunan & Alokasi Bulanan", "Buat satu baris untuk setiap pos anggaran. Isi input anggaran tahunan di sel kuning; phasing bulanan dihitung otomatis dari Asumsi.");
  s.getRange("A4:V4").merge(); s.getRange("A4:V4").values = [["Standar input: setiap pos harus punya Budget ID unik, owner yang akuntabel, kategori, dan nilai anggaran tahunan yang disetujui. Rencana bulanan mengikuti profil phasing perusahaan kecuali ada override resmi."]];
  s.getRange("A4:V4").format = { fill: lightTeal, font: { color: "#155E75", italic: true }, wrapText: true };
  const cols = ["Budget ID","Unit Bisnis","Cost Center","Layanan / Inisiatif","Driver COBIT","Jenis Belanja","Run / Change","Kategori Biaya","Owner Akuntabel","Anggaran Tahunan Disetujui","Jan","Feb","Mar","Apr","Mei","Jun","Jul","Agu","Sep","Okt","Nov","Des"];
  s.getRange("A6:V6").values = [cols]; header(s, "A6:V6"); s.getRange("A6:V6").format.rowHeight = 32;
  const seeds = [
    ["IT-001","TI Korporat","CC-IT-100","Infrastruktur Cloud","APO06","OpEx","Run","Cloud / Hosting","Infrastructure Manager",1200000000],
    ["IT-002","TI Korporat","CC-IT-200","Program Cybersecurity","APO06","OpEx","Change","Security","CISO",650000000],
    ["IT-003","Digital","CC-DIG-100","Modernisasi ERP","APO06","CapEx","Transform","Applications","CIO",1800000000],
    ["IT-004","TI Korporat","CC-IT-300","Layanan End User","APO06","OpEx","Run","Workplace & Devices","IT Operations Manager",480000000],
    ["IT-005","Digital","CC-DIG-200","Data & Analitik","APO06","CapEx","Change","Data Platform","CDO",900000000]
  ];
  const rows = Array.from({length:30}, (_,i) => i < seeds.length ? [...seeds[i], ...Array(12).fill(null)] : Array(22).fill(null));
  s.getRange("A7:V36").values = rows;
  inputStyle(s, "A7:J36"); formulaStyle(s, "K7:V36");
  s.getRange("J7:J36").format.numberFormat = money; s.getRange("K7:V36").format.numberFormat = money;
  s.getRange("K7").formulas = [["=IF($A7=\"\",\"\",ROUND($J7*'Asumsi'!F$6,0))"]]; s.getRange("K7:V7").fillRight(); s.getRange("K7:V36").fillDown();
  s.getRange("F7:F36").dataValidation = { rule: { type: "list", values: ["CapEx", "OpEx"] } };
  s.getRange("G7:G36").dataValidation = { rule: { type: "list", values: ["Run", "Change", "Transform"] } };
  s.getRange("A6:V36").format.borders = thin;
  s.tables.add("A6:V36", true, "BudgetInputTable").style = "TableStyleMedium2";
  s.getRange("A:V").format.columnWidth = 14;
  s.getRange("D:D").format.columnWidth = 28; s.getRange("H:H").format.columnWidth = 22; s.getRange("I:I").format.columnWidth = 25; s.getRange("J:V").format.columnWidth = 18;
  s.freezePanes.freezeRows(6); s.freezePanes.freezeColumns(4);
}

// BUDGET CONTROL
{
  const s = sheets["Kontrol Anggaran"];
  title(s, "A1:W1", "Anggaran TI — Kontrol Bulanan, Forecast & Manajemen Pengecualian", "Isi actual bulanan di sel kuning. Forecast dan varians otomatis mengikuti Bulan Pelaporan di Asumsi.");
  s.getRange("A4:W4").merge(); s.getRange("A4:W4").formulas = [["=\"Bulan pelaporan: \"&CHOOSE('Asumsi'!$B$6,\"Jan\",\"Feb\",\"Mar\",\"Apr\",\"Mei\",\"Jun\",\"Jul\",\"Agu\",\"Sep\",\"Okt\",\"Nov\",\"Des\")&\" | Varians positif = overspend | Kuning = input; biru = formula\""]];
  s.getRange("A4:W4").format = { fill: lightTeal, font: { color: "#155E75", italic: true } };
  const cols = ["Budget ID","Cost Center","Layanan / Inisiatif","Jenis Belanja","Kategori Biaya","Anggaran Tahunan","Actual Jan","Actual Feb","Actual Mar","Actual Apr","Actual Mei","Actual Jun","Actual Jul","Actual Agu","Actual Sep","Actual Okt","Actual Nov","Actual Des","Budget YTD","Actual YTD","Forecast FY","Varians","Varians %","Status Kontrol","Aksi / Komentar"];
  s.getRange("A6:Y6").values = [cols]; header(s, "A6:Y6"); s.getRange("A6:Y6").format.rowHeight = 36;
  s.getRange("A7:F36").formulas = Array.from({length:30},(_,i)=>{
    const r=i+7; return [`=IF('Input Anggaran'!A${r}=\"\",\"\",'Input Anggaran'!A${r})`,`=IF('Input Anggaran'!A${r}=\"\",\"\",'Input Anggaran'!C${r})`,`=IF('Input Anggaran'!A${r}=\"\",\"\",'Input Anggaran'!D${r})`,`=IF('Input Anggaran'!A${r}=\"\",\"\",'Input Anggaran'!F${r})`,`=IF('Input Anggaran'!A${r}=\"\",\"\",'Input Anggaran'!H${r})`,`=IF('Input Anggaran'!A${r}=\"\",\"\",'Input Anggaran'!J${r})`];
  });
  s.getRange("G7:R36").values = Array.from({length:30},()=>Array(12).fill(0));
  s.getRange("S7").formulas = [["=IF($A7=\"\",\"\",SUM('Input Anggaran'!$K7:INDEX('Input Anggaran'!$K7:$V7,1,'Asumsi'!$B$6)))"]];
  s.getRange("T7").formulas = [["=IF($A7=\"\",\"\",SUM($G7:INDEX($G7:$R7,1,'Asumsi'!$B$6)))"]];
  s.getRange("U7").formulas = [["=IF($A7=\"\",\"\",ROUND(SUM(IF('Asumsi'!$B$6>=1,$G7,'Input Anggaran'!$K7),IF('Asumsi'!$B$6>=2,$H7,'Input Anggaran'!$L7),IF('Asumsi'!$B$6>=3,$I7,'Input Anggaran'!$M7),IF('Asumsi'!$B$6>=4,$J7,'Input Anggaran'!$N7),IF('Asumsi'!$B$6>=5,$K7,'Input Anggaran'!$O7),IF('Asumsi'!$B$6>=6,$L7,'Input Anggaran'!$P7),IF('Asumsi'!$B$6>=7,$M7,'Input Anggaran'!$Q7),IF('Asumsi'!$B$6>=8,$N7,'Input Anggaran'!$R7),IF('Asumsi'!$B$6>=9,$O7,'Input Anggaran'!$S7),IF('Asumsi'!$B$6>=10,$P7,'Input Anggaran'!$T7),IF('Asumsi'!$B$6>=11,$Q7,'Input Anggaran'!$U7),IF('Asumsi'!$B$6>=12,$R7,'Input Anggaran'!$V7)),0))"]];
  s.getRange("V7").formulas = [["=IF($A7=\"\",\"\",$U7-$F7)"]];
  s.getRange("W7").formulas = [["=IF($A7=\"\",\"\",IFERROR($V7/$F7,0))"]];
  s.getRange("X7").formulas = [["=IF($A7=\"\",\"\",IF(ABS($W7)>='Asumsi'!$B$12,\"ESKALASI\",IF(ABS($W7)>='Asumsi'!$B$11,\"TINJAU\",\"TERKENDALI\")))"]];
  s.getRange("S7:Y36").fillDown();
  inputStyle(s, "G7:R36"); inputStyle(s, "Y7:Y36"); formulaStyle(s, "A7:F36"); formulaStyle(s, "S7:X36");
  s.getRange("F7:V36").format.numberFormat = money; s.getRange("W7:W36").format.numberFormat = "0.0%";
  s.getRange("A6:Y36").format.borders = thin;
  s.getRange("X7:X36").conditionalFormats.add("containsText", { text: "ESKALASI", format: { fill: "#F4CCCC", font: { bold: true, color: "#9C0006" } } });
  s.getRange("X7:X36").conditionalFormats.add("containsText", { text: "TINJAU", format: { fill: "#FFF2CC", font: { bold: true, color: "#9C6500" } } });
  s.getRange("X7:X36").conditionalFormats.add("containsText", { text: "TERKENDALI", format: { fill: "#D9EAD3", font: { color: "#274E13" } } });
  s.tables.add("A6:Y36", true, "BudgetControlTable").style = "TableStyleMedium2";
  s.getRange("A:Y").format.columnWidth = 15; s.getRange("C:C").format.columnWidth = 28; s.getRange("E:E").format.columnWidth = 22; s.getRange("F:X").format.columnWidth = 18; s.getRange("Y:Y").format.columnWidth = 36;
  s.freezePanes.freezeRows(6); s.freezePanes.freezeColumns(3);
}

// GOVERNANCE & APPROVALS
{
  const s = sheets["Tata Kelola & Persetujuan"];
  title(s, "A1:J1", "Log Tata Kelola, Persetujuan & Reforecast APO06", "Template mencakup ownership, bukti persetujuan, irama kerja, dan aktivitas kontrol keuangan yang selaras COBIT.");
  section(s, "A4:J4", "Register persetujuan / permintaan perubahan");
  const headers = ["Request / Versi","Tanggal Efektif","Cakupan / Budget ID","Jenis Perubahan","Nilai Diajukan","Business Case / Rasional","Disiapkan Oleh","Status Persetujuan","Approver / Ref Bukti","Tanggal Keputusan"];
  s.getRange("A5:J5").values = [headers]; header(s,"A5:J5");
  s.getRange("A6:J20").values = Array.from({length:15},(_,i)=> i===0 ? ["Anggaran Kerja v1.0",new Date("2027-01-01"),"Portofolio TI Tahunan","Persetujuan tahunan",5030000000,"Kerangka pendanaan tahunan, prioritisasi portofolio, dan alokasi biaya","Head of IT Finance","Menunggu","Catatan persetujuan CIO / CFO",null] : Array(10).fill(null));
  inputStyle(s,"A6:J20"); s.getRange("B6:B20").format.numberFormat="yyyy-mm-dd"; s.getRange("E6:E20").format.numberFormat=money; s.getRange("J6:J20").format.numberFormat="yyyy-mm-dd";
  s.getRange("H6:H20").dataValidation={rule:{type:"list",values:["Draft","Menunggu","Disetujui","Ditolak","Digantikan"]}};
  s.getRange("A5:J20").format.borders=thin; s.getRange("A:A").format.columnWidth=21; s.getRange("B:B").format.columnWidth=18; s.getRange("C:C").format.columnWidth=22; s.getRange("D:D").format.columnWidth=20; s.getRange("E:E").format.columnWidth=18; s.getRange("F:F").format.columnWidth=46; s.getRange("G:G").format.columnWidth=18; s.getRange("H:H").format.columnWidth=16; s.getRange("I:I").format.columnWidth=30; s.getRange("J:J").format.columnWidth=18;
  s.getRange("A6:J20").format.wrapText=true; s.getRange("A6:J20").format.rowHeight=30;
  section(s, "A23:J23", "Peta kontrol COBIT 2019 APO06 dan irama operasional");
  s.getRange("A24:F24").values = [["Fokus APO06","Aktivitas kontrol","Frekuensi","Akuntabel","Bukti","Tujuan kontrol"]]; header(s,"A24:F24");
  s.getRange("A25:F29").values = [
    ["Penetapan anggaran","Menetapkan anggaran TI tahunan dan periodik, sumber pendanaan, serta alokasi biaya","Tahunan + rolling forecast","CIO / IT Finance","Anggaran disetujui; planning pack","Menyelaraskan belanja dengan strategi dan portofolio perusahaan"],
    ["Manajemen biaya","Mencatat actual dan memantau varians budget-to-actual / forecast","Bulanan","IT Finance Manager","Tutup buku bulanan; laporan varians","Mengidentifikasi varians sejak dini dan memulai tindakan korektif"],
    ["Cost-benefit","Memvalidasi inisiatif material terhadap business case dan affordability","Saat intake dan stage gate","Portfolio Owner","Business case; persetujuan stage gate","Memastikan investasi dapat ditelusuri ke nilai yang diharapkan"],
    ["Reforecast","Memperbarui forecast, risiko, komitmen, dan action owner","Triwulanan atau berbasis trigger","Delegasi CIO / CFO","Versi reforecast; log persetujuan","Menjaga pendanaan yang prediktif dan berkelanjutan"],
    ["Chargeback / alokasi","Meninjau model alokasi dan mengkomunikasikan charge bila diterapkan","Triwulanan","IT Finance / Service Owner","Model alokasi; sign-off stakeholder","Memastikan alokasi biaya transparan dan terukur"]
  ];
  s.getRange("A24:F29").format.borders=thin; s.getRange("A25:F29").format.wrapText=true;
  s.getRange("A:A").format.columnWidth=22; s.getRange("B:B").format.columnWidth=42; s.getRange("C:C").format.columnWidth=22; s.getRange("D:D").format.columnWidth=26; s.getRange("E:E").format.columnWidth=30; s.getRange("F:F").format.columnWidth=44;
  s.freezePanes.freezeRows(5);
}

// CHECKS & SOURCES
{
  const s=sheets["Cek & Sumber"];
  title(s,"A1:H1","Cek Model & Register Sumber","Gunakan tab ini sebagai titik kontrol tutup buku bulanan dan bukti audit. LULUS berarti cek inti template saat ini rekonsiliasi.");
  section(s,"A4:F4","Status kontrol model");
  s.getRange("A5:D5").values=[["Cek","Hasil","Lokasi perbaikan","Catatan"]]; header(s,"A5:D5");
  s.getRange("A6:A10").values=[["Total phasing bulanan 100%"],["Baris input anggaran memiliki Budget ID"],["Forecast FY rekonsiliasi dengan kontrol anggaran"],["Tidak ada pengecualian forecast di atas ambang eskalasi"],["Status model"]];
  s.getRange("B6:B10").formulas=[
    ["=IF('Asumsi'!$Q$7=1,\"LULUS\",\"GAGAL\")"],
    ["=IF(COUNTBLANK('Input Anggaran'!$A$7:$A$11)=0,\"LULUS\",\"TINJAU\")"],
    ["=IF(ABS(SUM('Kontrol Anggaran'!$U$7:$U$36)-SUM('Kontrol Anggaran'!$F$7:$F$36)-SUM('Kontrol Anggaran'!$V$7:$V$36))<1,\"LULUS\",\"GAGAL\")"],
    ["=IF(COUNTIF('Kontrol Anggaran'!$X$7:$X$36,\"ESKALASI\")=0,\"LULUS\",\"TINJAU\")"],
    ["=IF(COUNTIF(B6:B9,\"GAGAL\")>0,\"GAGAL\",IF(COUNTIF(B6:B9,\"TINJAU\")>0,\"TINJAU\",\"LULUS\"))"]
  ];
  s.getRange("C6:C10").values=[["Asumsi!F6:Q6"],["Input Anggaran!A7:A36"],["Kontrol Anggaran!F:V"],["Kontrol Anggaran!X:X"],["Selesaikan item GAGAL atau TINJAU di atas"]];
  s.getRange("D6:D10").values=[["Distribusi harus tepat 100%."],["Minimal baris anggaran aktif wajib memiliki ID."],["Forecast = budget + varians."],["Eskalasi varians mengikuti ambang batas pada Asumsi."],["Ringkasan tindakan kontrol yang dibutuhkan."]];
  s.getRange("A5:D10").format.borders=thin; formulaStyle(s,"B6:B10"); s.getRange("A:A").format.columnWidth=38; s.getRange("B:B").format.columnWidth=16; s.getRange("C:C").format.columnWidth=34; s.getRange("D:D").format.columnWidth=70;
  s.getRange("A6:D10").format.wrapText=true; s.getRange("A6:D10").format.rowHeight=28;
  s.getRange("B6:B10").conditionalFormats.add("containsText",{text:"GAGAL",format:{fill:"#F4CCCC",font:{bold:true,color:"#9C0006"}}});
  s.getRange("B6:B10").conditionalFormats.add("containsText",{text:"TINJAU",format:{fill:"#FFF2CC",font:{bold:true,color:"#9C6500"}}});
  s.getRange("B6:B10").conditionalFormats.add("containsText",{text:"LULUS",format:{fill:"#D9EAD3",font:{bold:true,color:"#274E13"}}});
  section(s,"A13:J13","Register sumber");
  s.getRange("A14:J14").values=[["Item","Nilai / Deskripsi","Unit","Periode / As-of","Jenis Sumber","Nama Sumber","URL Referensi","Owner","Catatan","Diakses / Diperbarui"]]; header(s,"A14:J14");
  s.getRange("A15:J17").values=[
    ["Framework tata kelola","Referensi COBIT 2019 governance and management objective","Framework","2019","Standar eksternal","ISACA COBIT resource center","https://www.isaca.org/resources/cobit","IT Governance","APO06 digunakan sebagai anchor desain tata kelola finansial.",new Date("2026-09-01")],
    ["Konteks APO06","Managed Budget and Costs; pemilihan target capability perlu disesuaikan dengan design factor perusahaan","Referensi framework","2019","Standar eksternal","Artikel ISACA","https://www.isaca.org/resources/news-and-trends/industry-news/2019/defining-target-capability-levels-in-cobit-2019-a-proposal-for-refinement","IT Governance","Konfirmasi praktik proses spesifik organisasi dengan konten COBIT 2019 Governance and Management Objectives berlisensi.",new Date("2026-09-01")],
    ["Input anggaran","Rencana tahunan/periodik, actual, komitmen, dan forecast","IDR","Tahun Fiskal","Input internal","ERP / GL / procurement / sistem portofolio","","IT Finance","Ganti contoh dengan data organisasi yang sudah disetujui dan simpan ID laporan sumber.",new Date("2026-09-01")]
  ];
  s.getRange("A14:J17").format.borders=thin; s.getRange("A:A").format.columnWidth=20; s.getRange("B:B").format.columnWidth=36; s.getRange("C:C").format.columnWidth=16; s.getRange("D:D").format.columnWidth=18; s.getRange("E:E").format.columnWidth=18; s.getRange("F:F").format.columnWidth=24; s.getRange("G:G").format.columnWidth=50; s.getRange("H:H").format.columnWidth=18; s.getRange("I:I").format.columnWidth=52; s.getRange("J:J").format.columnWidth=18; s.getRange("J15:J17").format.numberFormat="yyyy-mm-dd"; s.getRange("A15:J17").format.wrapText=true; s.getRange("A15:J17").format.rowHeight=52;
  s.freezePanes.freezeRows(5);
}

// DASHBOARD
{
  const s=sheets["Dasbor"];
  title(s,"A1:P1","Dasbor Eksekutif Anggaran TI — APO06 Managed Budget and Costs","Tampilan tata kelola tahunan/periodik. Ubah bulan pelaporan dan threshold di Asumsi; isi actual di Kontrol Anggaran.");
  s.getRange("A4:B4").merge(); s.getRange("A4:B4").values=[["Tahun Fiskal"]]; s.getRange("C4:D4").merge(); s.getRange("C4:D4").formulas=[["='Asumsi'!$B$5"]];
  s.getRange("E4:F4").merge(); s.getRange("E4:F4").values=[["Bulan Pelaporan"]]; s.getRange("G4:H4").merge(); s.getRange("G4:H4").formulas=[["=CHOOSE('Asumsi'!$B$6,\"Jan\",\"Feb\",\"Mar\",\"Apr\",\"Mei\",\"Jun\",\"Jul\",\"Agu\",\"Sep\",\"Okt\",\"Nov\",\"Des\")"]];
  s.getRange("I4:J4").merge(); s.getRange("I4:J4").values=[["Status Model"]]; s.getRange("K4:L4").merge(); s.getRange("K4:L4").formulas=[["='Cek & Sumber'!$B$10"]];
  s.getRange("A4:L4").format={fill:lightGray,font:{bold:true},horizontalAlignment:"center",borders:thin}; s.getRange("C4:D4").format={fill:lightBlue,font:{bold:true,size:12},horizontalAlignment:"center",borders:thin}; s.getRange("G4:H4").format={fill:lightBlue,font:{bold:true,size:12},horizontalAlignment:"center",borders:thin}; formulaStyle(s,"K4:L4");
  const cards=[["Anggaran Tahunan Disetujui","=SUM('Kontrol Anggaran'!$F$7:$F$36)"],["Budget YTD","=SUM('Kontrol Anggaran'!$S$7:$S$36)"],["Actual YTD","=SUM('Kontrol Anggaran'!$T$7:$T$36)"],["Forecast FY","=SUM('Kontrol Anggaran'!$U$7:$U$36)"],["Varians Forecast","=SUM('Kontrol Anggaran'!$V$7:$V$36)"],["Eskalasi","=COUNTIF('Kontrol Anggaran'!$X$7:$X$36,\"ESKALASI\")"]];
  const starts=["A6","D6","G6","J6","M6","P6"];
  for(let i=0;i<cards.length;i++){ const col=starts[i].replace(/\d/,""); const start=starts[i]; const end=String.fromCharCode(col.charCodeAt(0)+1)+"8"; const r=`${start}:${end}`; const expr=cards[i][1].replace(/^=/,""); s.getRange(r).merge(); s.getRange(r).format={fill:i===4?red:(i===5?amber:lightBlue),borders:{preset:"outside",style:"medium",color:blue},horizontalAlignment:"center",verticalAlignment:"center",wrapText:true}; s.getRange(start).formulas=[[`=\"${cards[i][0]}\"&CHAR(10)&TEXT(${expr},\"#,##0\")`]]; s.getRange(start).format.font={bold:true,color:navy,size:11}; }
  // Data bantu berbasis formula untuk grafik tren.
  s.getRange("A12:D12").values=[["Bulan","Rencana Budget","Actual","Varians"]]; header(s,"A12:D12");
  s.getRange("A13:A24").values=months.map(x=>[x]);
  for(let i=0;i<12;i++){const r=i+13, col=String.fromCharCode(75+i); const actualCol=String.fromCharCode(71+i); s.getRange(`B${r}`).formulas=[[`=SUM('Input Anggaran'!$${col}$7:$${col}$36)`]]; s.getRange(`C${r}`).formulas=[[`=SUM('Kontrol Anggaran'!$${actualCol}$7:$${actualCol}$36)`]]; s.getRange(`D${r}`).formulas=[[`=C${r}-B${r}`]];}
  s.getRange("B13:D24").format.numberFormat=money; s.getRange("A12:D24").format.borders=thin;
  const chart=s.charts.add("line",s.getRange("A12:C24")); chart.title="Rencana Budget Bulanan vs Actual (IDR)"; chart.hasLegend=true; chart.xAxis={axisType:"textAxis"}; chart.yAxis={numberFormatCode:"#,##0"}; chart.setPosition("F11","P28");
  section(s,"A27:P27","Pos anggaran yang membutuhkan tindakan manajemen");
  s.getRange("A28:F28").values=[["Budget ID","Inisiatif","Forecast FY","Varians","Varians %","Status Kontrol"]]; header(s,"A28:F28");
  for(let i=0;i<5;i++){const r=i+29, source=i+7; s.getRange(`A${r}:F${r}`).formulas=[[`='Kontrol Anggaran'!A${source}`,`='Kontrol Anggaran'!C${source}`,`='Kontrol Anggaran'!U${source}`,`='Kontrol Anggaran'!V${source}`,`='Kontrol Anggaran'!W${source}`,`='Kontrol Anggaran'!X${source}`]];}
  s.getRange("C29:D33").format.numberFormat=money;s.getRange("E29:E33").format.numberFormat="0.0%";s.getRange("A28:F33").format.borders=thin;
  s.getRange("F29:F33").conditionalFormats.add("containsText",{text:"ESKALASI",format:{fill:"#F4CCCC",font:{bold:true,color:"#9C0006"}}});
  s.getRange("F29:F33").conditionalFormats.add("containsText",{text:"TINJAU",format:{fill:"#FFF2CC",font:{bold:true,color:"#9C6500"}}});
  s.getRange("A:P").format.columnWidth=14;s.getRange("B:B").format.columnWidth=24;s.getRange("A:A").format.columnWidth=16;s.getRange("F:F").format.columnWidth=22;
  s.freezePanes.freezeRows(4);
}

await fs.mkdir(outputDir, { recursive: true });
for (const sheet of Object.values(sheets)) { try { sheet.getUsedRange().format.autofitRows(); } catch {} }
const output = await SpreadsheetFile.exportXlsx(wb);
await output.save(`${outputDir}/Template_Anggaran_TI_APO06_COBIT2019_INDONESIA.xlsx`);

const check = await wb.inspect({kind:"table",range:"Dasbor!A1:P33",include:"values,formulas",tableMaxRows:33,tableMaxCols:16});
console.log(check.ndjson);
const errors=await wb.inspect({kind:"match",searchTerm:"#REF!|#DIV/0!|#VALUE!|#NAME\\?|#N/A",options:{useRegex:true,maxResults:100},summary:"final formula error scan"});
console.log(errors.ndjson);
for (const name of names) { const image=await wb.render({sheetName:name,autoCrop:"all",scale:1,format:"png"}); await fs.writeFile(`${outputDir}/${name.replace(/[^A-Za-z0-9]/g,"_")}.png`,new Uint8Array(await image.arrayBuffer())); }

