const fs = require('fs');
//const Influx = require('influx');
//const constants = require('./constants.js');
const exec = require('child_process').exec;
//const request = require('request');


/** constants with providers, languages and tests */
const AWS = 'aws';
const AZURE = 'azure';
const AZUREWINDOWS = 'azureWindows';
const GOOGLE = 'google';
const IBM = 'ibm';

const NODE = 'node';
const PYTHON = 'python';
const GO = 'go';
const DOTNET = 'dotnet';

const LATENCY = 'latency';
const FACTORS = 'factors';
const MATRIX = 'matrix';
const FILESYSTEM = 'filesystem';
const CUSTOM = 'custom';

const PROVIDERS = [AWS, AZURE, AZUREWINDOWS, GOOGLE, IBM];
const LANGUAGES = [NODE, PYTHON, GO, DOTNET];
const TESTS = [LATENCY, FACTORS, MATRIX, FILESYSTEM, CUSTOM];


const tags = ['test', 'provider', 'language', 'memory'];

/* const influx = new Influx.InfluxDB({
    host: 'localhost',
    port: 8086,
    database: 'results',
    username: 'benchmark-suite',
    password: 'benchmark',
    schema: [
        {
            measurement: constants.LATENCY+'_benchmark',
            fields: {
                duration: Influx.FieldType.INTEGER,
                latency_avg: Influx.FieldType.FLOAT,
                latency_stdev: Influx.FieldType.FLOAT,
                latency_max: Influx.FieldType.FLOAT,
                total_count: Influx.FieldType.INTEGER,
                rps_theoretic: Influx.FieldType.INTEGER,
                rps_avg: Influx.FieldType.FLOAT,
                percentile_50: Influx.FieldType.FLOAT,
                percentile_95: Influx.FieldType.FLOAT,
                percentile_99: Influx.FieldType.FLOAT,
                error: Influx.FieldType.INTEGER,
                n: Influx.FieldType.INTEGER
            },
            tags: tags
        },
        {
            measurement: constants.FACTORS+'_benchmark',
            fields: {
                duration: Influx.FieldType.INTEGER,
                latency_avg: Influx.FieldType.FLOAT,
                latency_stdev: Influx.FieldType.FLOAT,
                latency_max: Influx.FieldType.FLOAT,
                total_count: Influx.FieldType.INTEGER,
                rps_theoretic: Influx.FieldType.INTEGER,
                rps_avg: Influx.FieldType.FLOAT,
                percentile_50: Influx.FieldType.FLOAT,
                percentile_95: Influx.FieldType.FLOAT,
                percentile_99: Influx.FieldType.FLOAT,
                error: Influx.FieldType.INTEGER,
                n: Influx.FieldType.INTEGER
            },
            tags: tags
        },
        {
            measurement: constants.MATRIX+'_benchmark',
            fields: {
                duration: Influx.FieldType.INTEGER,
                latency_avg: Influx.FieldType.FLOAT,
                latency_stdev: Influx.FieldType.FLOAT,
                latency_max: Influx.FieldType.FLOAT,
                total_count: Influx.FieldType.INTEGER,
                rps_theoretic: Influx.FieldType.INTEGER,
                rps_avg: Influx.FieldType.FLOAT,
                percentile_50: Influx.FieldType.FLOAT,
                percentile_95: Influx.FieldType.FLOAT,
                percentile_99: Influx.FieldType.FLOAT,
                error: Influx.FieldType.INTEGER,
                n: Influx.FieldType.INTEGER
            },
            tags: tags
        },
        {
            measurement: constants.CUSTOM+'_benchmark',
            fields: {
                duration: Influx.FieldType.INTEGER,
                latency_avg: Influx.FieldType.FLOAT,
                latency_stdev: Influx.FieldType.FLOAT,
                latency_max: Influx.FieldType.FLOAT,
                total_count: Influx.FieldType.INTEGER,
                rps_theoretic: Influx.FieldType.INTEGER,
                rps_avg: Influx.FieldType.FLOAT,
                percentile_50: Influx.FieldType.FLOAT,
                percentile_95: Influx.FieldType.FLOAT,
                percentile_99: Influx.FieldType.FLOAT,
                error: Influx.FieldType.INTEGER,
                n: Influx.FieldType.INTEGER
            },
            tags: tags
        }
    ]
});
 */
var allUrls = {
    latency: ["http://host.docker.internal:8081/"],
    factors: ["http://host.docker.internal:3000/matrix"],
    matrix:  ["http://host.docker.internal:3000/latency"]
};


loadtest(FACTORS, "mytest", 1, 80, 1);

/** Executes a shell command and return it as a Promise. */
function execShellCommand(cmd) {
	return new Promise((resolve, reject) => {
		exec(cmd, (error, stdout, stderr) => {
	  		if (error) {
				let errorMsg = 'TIME: ' + (new Date(Date.now() - ((new Date()).getTimezoneOffset() * 60000))).toISOString() + '\n';
				errorMsg += 'ERROR: ' + error;
				errorMsg += 'STDOUT: ' + stdout + (stdout? '' : '\n');
				errorMsg += 'STDERR: ' + stderr + (stderr? '' : '\n');
				errorMsg += '----------------------------------------------------------------------------------------------------\n';
				reject(errorMsg);
	  		}		
	  		resolve(stdout? stdout : stderr);
	 	});
	});
}

async function loadtest(test, testName, rps, duration, n) {

    const wrk2 = 'docker run --rm bschitter/alpine-with-wrk2:0.1';

    let error = false;

        if(duration<30) {
            //duration = 30;
        }
 /*        if(test == constants.FACTORS || test == constants.MATRIX) {
            allUrls[test][i].url += '?n=' + n;
        } else {
            n = 0;
        }
 */
        // 1 request until completed
        //await request.get(allUrls[test][0]);

        // run for 10 seconds to avoid cold start latencies
        await execShellCommand(wrk2 + ' -c100 -t2 -d10s -R' + rps + ' -L ' + allUrls[test][0]);        

        // sleep 10 seconds to avoid that the function is still under load
        //await new Promise(resolve => setTimeout(resolve, 10000));

        // benchmark
        let result = await execShellCommand(wrk2 + ' -c20 -t2 -d' + duration + 's -R' + rps + ' -L ' + allUrls[test][0])
        .catch((err) => {
            console.error("the error", err);
            error = true;
        });

        

        if(error) {
            return;
        }

        let latency_avg = 0;
        let latency_stdev = 0;
        let latency_max = 0;
        let total_count = 0;
        let rps_avg = 0;
        let percentile_50 = 0;
        let percentile_95 = 0;
        let percentile_99 = 0;
        let errors = 0;

        console.log(result)

        let lines = result.split('\n');
        for(let j=0; j<lines.length; j++) {
            if(lines[j].startsWith('#[Mean')) {
                let cleanLine = lines[j].replace(/ /g, '');
                let parts = cleanLine.split(',');
                latency_avg = parts[0].substring(parts[0].indexOf("=") + 1);
                latency_stdev = parts[1].substring(parts[1].indexOf("=") + 1).replace(']','');

            } else if(lines[j].startsWith('#[Max')) {
                let cleanLine = lines[j].replace(/ /g, '');
                let parts = cleanLine.split(',');
                latency_max = parts[0].substring(parts[0].indexOf("=") + 1);
                total_count = parts[1].substring(parts[1].indexOf("=") + 1).replace(']','');

            } else if(lines[j].startsWith('Requests/sec:')) {
                let cleanLine = lines[j].replace(/ /g, '');
                rps_avg = cleanLine.substring(cleanLine.indexOf(":") + 1);

            } else if(lines[j].startsWith('  Socket errors:')) {
                let cleanLine = lines[j].replace(/ /g, '');
                cleanLine = cleanLine.substring(cleanLine.indexOf(":") + 1);
                let parts = cleanLine.split(',');
                for(let k=0; k<parts.length; k++) {
                    errors += Number(parts[k].replace(/([^0-9])/g,''));
                }

            } else if(lines[j].startsWith(' 50.000%')) {
                let cleanLine = lines[j].replace(/ /g, '');
                if(cleanLine.includes('ms')) {
                    cleanLine = cleanLine.replace(/([a-z])/g,'');
                    percentile_50 = cleanLine.substring(cleanLine.indexOf("%") + 1);
                } else {
                    cleanLine = cleanLine.replace(/([a-z])/g,'');
                    percentile_50 = cleanLine.substring(cleanLine.indexOf("%") + 1) * 1000;
                }

            } else if(lines[j].startsWith(' 99.000%')) {
                let cleanLine = lines[j].replace(/ /g, '');
                if(cleanLine.includes('ms')) {
                    cleanLine = cleanLine.replace(/([a-z])/g,'');
                    percentile_99 = cleanLine.substring(cleanLine.indexOf("%") + 1);
                } else {
                    cleanLine = cleanLine.replace(/([a-z])/g,'');
                    percentile_99 = cleanLine.substring(cleanLine.indexOf("%") + 1) * 1000;
                }

            } else if(lines[j].includes('0.950000')) {
                let cleanLine = lines[j].replace(/  +/g, ' ');
                let parts = cleanLine.split(' ');
                percentile_95 = Math.round(parts[1] * 100) / 100;
            }
        

        //insertIntoDB(test+'_benchmark', testName, allUrls[test][i].language, allUrls[test][i].provider, allUrls[test][i].memory, duration, latency_avg, latency_stdev, latency_max, total_count, rps, rps_avg, percentile_50, percentile_95, percentile_99, errors, n);

        var data = [
            {
              measurement: test,
              tags: {
                test: testName,
                language: "go",
                provider: "local",
                memory: 20
              },
              fields: {
                duration: duration,
                latency_avg: latency_avg,
                latency_stdev: latency_stdev,
                latency_max: latency_max,
                total_count: total_count,
                rps_theoretic: rps,
                rps_avg: rps_avg,
                percentile_50: percentile_50,
                percentile_95: percentile_95,
                percentile_99: percentile_99,
                error: errors,
                n: n
              },
            }
        ];

        //console.log(data)

    }

    return !error;
}

/* function insertIntoDB(test, testName, language, provider, memory, duration, latency_avg, latency_stdev, latency_max, total_count, rps, rps_avg, percentile_50, percentile_95, percentile_99, errors, n) {

    var data = [
        {
          measurement: test,
          tags: {
            test: testName,
            language: language,
            provider: provider,
            memory: memory
          },
          fields: {
            duration: duration,
            latency_avg: latency_avg,
            latency_stdev: latency_stdev,
            latency_max: latency_max,
            total_count: total_count,
            rps_theoretic: rps,
            rps_avg: rps_avg,
            percentile_50: percentile_50,
            percentile_95: percentile_95,
            percentile_99: percentile_99,
            error: errors,
            n: n
          },
        }
    ];

    influx.writePoints(data)
    .catch((err) => {
        console.error(err);
        console.error('Could not write to DB!');
    });
} */




module.exports = { loadtest };