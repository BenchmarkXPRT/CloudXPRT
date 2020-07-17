var username = window.localStorage.getItem("username")

function logout() {
    localStorage.removeItem("username");
    window.location = "index.html"
}

async function runMC() {
    // Call REST API for running MonteCarlo
    url = 'http://' + location.host + '/mc?name=' + username
    const response = await fetch(url, {
        method: "GET",
        headers: {
            "Content-Type": "application/json"
        }
    })
    .then(response => response.json())
    .then(data => {
        var table = new Tabulator("#options-table", {
            data:data,
            layout:"fitColumns",
            pagination:"local",
            paginationSize:10,
            paginationSizeSelector:[5,10],
            movableColumns:true,
            columns:[
            {title:"Stock Price", field:"stockprice"},
            {title:"Option Strike Price", field:"strikeprice"},
            {title:"Option Years", field:"optionyear"},
            {title:"Call Result", field:"callresult"}
            //{title:"Call Confidence", field:"callconfidence"},
            ],
        });
    });
}

// Tabulator parameters

//var table = new Tabulator("#example-table", {
//	data:tabledata,           //load row data from array
//	layout:"fitColumns",      //fit columns to width of table
//	responsiveLayout:"hide",  //hide columns that dont fit on the table
//	tooltips:true,            //show tool tips on cells
//	addRowPos:"top",          //when adding a new row, add it to the top of the table
//	history:true,             //allow undo and redo actions on the table
//	pagination:"local",       //paginate the data
//	paginationSize:7,         //allow 7 rows per page of data
//	movableColumns:true,      //allow column order to be changed
//	resizableRows:true,       //allow row order to be changed
//	initialSort:[             //set the initial sort order of the data
//		{column:"name", dir:"asc"},
//	],
//	columns:[                 //define the table columns
//		{title:"Name", field:"name", editor:"input"},
//		{title:"Task Progress", field:"progress", align:"left", formatter:"progress", editor:true},
//		{title:"Gender", field:"gender", width:95, editor:"select", editorParams:{values:["male", "female"]}},
//		{title:"Rating", field:"rating", formatter:"star", align:"center", width:100, editor:true},
//		{title:"Color", field:"col", width:130, editor:"input"},
//		{title:"Date Of Birth", field:"dob", width:130, sorter:"date", align:"center"},
//		{title:"Driver", field:"car", width:90,  align:"center", formatter:"tickCross", sorter:"boolean", editor:true},
//	],
//});
