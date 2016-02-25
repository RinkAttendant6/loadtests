'use strict';

const argv = require('minimist')(process.argv.slice(2));
const inquirer = require('inquirer');
const randomstring = require('randomstring');
const Client = require('./Client');

const host = argv.host || process.env.LOADTESTSME_SCHEDULER_HOST;
const port = argv.port || process.env.LOADTESTSME_SCHEDULER_PORT || 8080;

const client = new Client(host, port);

inquirer.prompt([
    {
        type: 'input',
        name: 'script_name',
        message: 'Name of test (for identification purposes):',
        default: randomstring.generate({
            length: 32,
            charset: 'hex'
        })
    },
    {
        type: 'input',
        name: 'run_time',
        message: 'Run time (seconds):',
        default: 10,
        validate: input => !isNaN(Number(input))
    },
    {
        type: 'input',
        name: 'growth_factor',
        message: 'Growth factor:',
        default: 1,
        validate: input => !isNaN(Number(input))
    },
    {
        type: 'input',
        name: 'starting_requests_per_second',
        message: 'Starting requests per second:',
        default: 15,
        validate: input => !isNaN(Number(input))
    },
    {
        type: 'input',
        name: 'max_requests_per_second',
        message: 'Maximum requests per second:',
        default: 75,
        validate: input => !isNaN(Number(input))
    }
], answers => {
    console.log(answers);
    
    let loadTest = client.createLoadTest(answers);
    console.log(loadTest);
});