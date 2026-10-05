class StorageObjeto {
    constructor(nome){
        this.nome = nome
    }
    save(objeto){
        localStorage.setItem(this.nome, JSON.stringify(objeto))
    }

    get(){
        // Pega o objeto em string no localstorage
        let objetoString = localStorage.getItem(this.nome)
    
        // verifca se o local storage esta vazio
        if (objetoString == '') {
            return 
        }
    
        // transformar em objeto novamente
        let objeto = JSON.parse(objetoString)
        return objeto
    }

    clear(){
        localStorage.setItem(this.nome, '')
    }

    verificar(){
        if (localStorage.getItem(this.nome) == ""){
            return false
        }else {
            return true
        }
    }
}