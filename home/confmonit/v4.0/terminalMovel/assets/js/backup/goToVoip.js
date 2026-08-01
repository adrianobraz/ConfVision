function goToLigar(){
    

    $.ajax({
        url: 'https://api.goto.com/calls/v2/calls',
        method: 'Post',
        // headers: {
        //     "Content-Type": "application/json",
        //     "Accept": "application/json",
        //     "Authorization": "Bearer " + sessionStorage.getItem('token')
        // },
        data: JSON.stringify({
            dialString: '(866) 768-5429',
            from: {lineId: '014ce388-da2c-d120-88f9-000100320002'},
            autoAnswer: false,
            phoneNumberId: '8e9ce3fb-4d22-4fe3-baa7-b7892dd942d7'
          })
    }).fail(function (e) {
        console.log(e)
    }).done(function (r) {
console.log(r)
    })


}