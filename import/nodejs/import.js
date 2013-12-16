var http = require('http'),
	importBase = "/api/wow/item/",
	startTime = process.hrtime(),
	printDuration = function(time) {
		return process.hrtime(time)[1] / 1000000;
	},
	startIndex = 1000,
	endIndex = 80000
	;

var printError = function(d) {
	console.log('ERROR ' + d);
};

var nullFunc = function(d) {
	return;
}

var saveDoc = function(index, type, key, json) {
	if (key == undefined) {
		console.log('Skipping');
		return;
	}
	var options = {
		hostname: 'localhost',/*'192.168.1.159',*/
		port: 9200,
		path: index + '/' + type + '/' + key,
		method: 'PUT'
	}

	var req = http.request(options, function(res) {
		res.on('data', nullFunc);
		res.on('end', nullFunc);
	});
	req.on('error', printError);
	req.write(json);
	req.end();
	console.log('PUT document to ' + options.path)
}

for (var i = startIndex; i < endIndex; i++) {
	var options = {
		hostname: 'eu.battle.net',
		port: 80,
		path: importBase + i,
		method: 'GET'
	};

	var req = http.request(options, function(res) {
		// console.log('Received ' + res.statusCode);
		res.setEncoding('utf8');
		var respBody = '';
		res.on('data', function(d) {
			respBody += d;
		})

		res.on('end', function(d) {
			var obj = JSON.parse(respBody);
			// console.log('Read WoW Item ' + obj.id 
			// 	+ ' with name: ' + obj.name 
			// 	+ ' in ' + printDuration(startTime) + 'ms');
			if (obj == undefined || obj.id == undefined) {
				return;
			}

			console.log('Saving WoW Item to ElasticSearch with name ' + obj.name);

			saveDoc('wowstatic', 'item', obj.id, respBody);

		});

		res.on('error', printError);
	});
	req.on('error', printError);
	req.end();
}

console.log('Done in ' + printDuration(startTime));
