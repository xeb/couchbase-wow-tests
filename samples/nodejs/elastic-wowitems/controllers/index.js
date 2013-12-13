'use strict';


module.exports = function (server) {

    server.get('/', function (req, res) {
        var model = { name: 'I am a test' };
        
        res.render('index', model);
        
    });

    
    server.get('/indexing', function (req, res) {
        
        res.render('indexing');
        
    });

};
