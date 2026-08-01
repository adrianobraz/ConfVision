function gerarPdf(titulo, tabela, orientacao = 'portrait', fechamento = "") {
   
    // importa biblioteca
    var jsPDF = window.jspdf.jsPDF
    // instancia a biblioteca
    var doc = new jsPDF({
        orientation: orientacao,
        unit: 'pt',
    })
    var totalPagesExp = '{total_pages_count_string}'

    doc.autoTable({
        html: tabela,
        theme: 'striped',  
        // formata o hed da tabela
        headStyles: { 
            textColor: 'White',
            lineColor: [128,128,128], 
            fillColor: [105,105,105], 
            halign: 'center', 
            lineWidth: 2, 
        },
        bodyStyles: { lineColor: [128,128,128], lineWidth: 1, textColor: 'Black'},

        //startY: 70, // posicao do inicio da tabela na primeira pagia
        didDrawPage: function (data) {
            // imprime o numero da pagina
            var pageSize = doc.internal.pageSize

            // Header
            doc.setFontSize(15)
            doc.setTextColor(40)
            doc.text(titulo, calculateTextWidth(pageSize.getWidth(), titulo), 30)
            doc.setFontSize(10)
            doc.text(dataAtualFormatada(new Date()), pageSize.getWidth() - 130, 30)

            // Footer
            var str = 'Page ' + doc.internal.getNumberOfPages()
            // Total page number plugin only available in jspdf v1.0+
            if (typeof doc.putTotalPages === 'function') {
                str = str + ' of ' + totalPagesExp
            }

            doc.setFontSize(10)
            doc.text(str, data.settings.margin.left, pageSize.getHeight() - 10)
            doc.text('confmonit', pageSize.getWidth() - 100, pageSize.getHeight() - 10)
        },


    })
    if (fechamento != '') {
        doc.text(fechamento, 35, doc.lastAutoTable.finalY + 30)
    }

    // Total page number plugin only available in jspdf v1.0+
    if (typeof doc.putTotalPages === 'function') {
        doc.putTotalPages(totalPagesExp)
    }


    //Abre a janela de visualização e impressao
    window.open(doc.output('bloburl'), '_blank')
    doc.autoPrint()
    doc.output("dataurlnewwindows")
}


var calculateTextWidth = function (tmanhoPagina, text) {
    const span = document.createElement('span');
    span.innerText = text;

    document.body.appendChild(span);
    const tamanhoTexto = span.offsetWidth * 1.3281472327365
    span.parentNode.removeChild(span);

    return (tmanhoPagina - tamanhoTexto) / 2
}

function dataAtualFormatada(data){
    var data = new Date(data),
        dia  = data.getDate().toString(),
        diaF = (dia.length == 1) ? '0'+dia : dia,
        mes  = (data.getMonth()+1).toString(), //+1 pois no getMonth Janeiro começa com zero.
        mesF = (mes.length == 1) ? '0'+mes : mes,
        anoF = data.getFullYear();
    return diaF+"/"+mesF+"/"+anoF+" "+data.getHours()+":"+data.getMinutes()+":"+data.getSeconds()
}
