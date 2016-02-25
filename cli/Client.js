'use strict';

const grpc = require('grpc');
const path = require('path');
const scheduler = grpc.load(path.join(__dirname, '..', 'pb', 'scheduler.proto')).loadtests;
const sprintf = require('sprintf-js').sprintf;

/**
 * Class representing a command-line client 
 */
class Client {
    /**
     * Creates a new client
     * @param {string} host - The hostname (IP address or domain name) of the scheduler
     * @param {number} port - The port number of the scheduler
     */
    constructor(host, port) {
        this.jobs = new Map();
        this.scheduler = new scheduler.Scheduler(sprintf('%s:%d', host, port), grpc.credentials.createInsecure());
    }
    
    /**
     * Runs a load test
     * @param {Object} req - Load test request
     * @param {number} testID - An ID given to identify the load test
     * @returns {Object}
     */
    runLoadTest(req, testID) {
        req.set_script_name(String(testID));

        let job = this.scheduler.loadTest(req);

        // Put job into collection of jobs to allow for stopping
        this.jobs.set(testID, job);

        // Remove from collection when done
        job.on('end', () => {
            this.jobs.delete(testID);
        });

        // Return a reference to it
        return job;
    }
    
    /**
     * Creates a new load test
     * @param {Object.<string, *>} properties - Load test parameters
     * @returns {Object}
     */
    createLoadTest(properties) {
        console.log(this.scheduler);
        let request = new this.scheduler.LoadTestReq();
        Object.keys(properties)
            .filter(k => request[`set_${k}`] !== undefined)
            .map(k => {
                request[`set_${k}`](properties[k]);
            });
            
        return request;
    }
}

module.exports = Client;