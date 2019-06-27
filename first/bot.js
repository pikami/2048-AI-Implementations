const reimprove = require('reimprovejs/dist/reimprove.js');
const { remote } = require('webdriverio');

(async() => {
    // Setup browser
    const browser = await remote({
        logLevel: 'error',
        path: '/',
        port: 9515,
        capabilities: {
            browserName: 'chrome'
        }
    }); 
    browser.url('https://4ark.me/2048/');
    await browser.pause(1000);
    const resetButton = await browser.$('.restart-btn');
    const failMessage = await browser.$('.failure-container');

    // Function to reset game
    async function resetGame(){
        steps = 0;
        gen++;
        await resetButton.click();
        await browser.pause(50);
    }

    // A.I. starts here
    let steps = 0;
    let gen = 0;
    let stepsWithoutGain = 0;
    const modelFitConfig = {
        epochs: 12,
        stepsPerEpoch: 16
    };

    const numActions = 4;
    const inputSize = 16;
    // The window of data which will be sent yo your agent. For instance the x previous inputs, and what actions the agent took  
    const temporalWindow = 16;

    const totalInputSize = inputSize * temporalWindow + numActions * temporalWindow + inputSize;

    const network = new reimprove.NeuralNetwork();
    network.InputShape = [totalInputSize];
    network.addNeuralNetworkLayers([
        {type: 'dense', units: 32, activation: 'relu'},
        {type: 'dense', units: 32, activation: 'relu'},
        {type: 'dense', units: numActions, activation: 'softmax'}
    ]);
    // Now we initialize our model, and start adding layers
    const model = new reimprove.Model.FromNetwork(network, modelFitConfig);

    // Finally compile the model, we also exactly use tfjs's optimizers and loss functions
    // (So feel free to choose one among tfjs's)
    model.compile({loss: 'meanSquaredError', optimizer: 'sgd'})

    // Every single field here is optionnal, and has a default value. Be careful, it may not fit your needs ...
    const teacherConfig = {
        lessonsQuantity: 10000,
        lessonLength: 500,                
        lessonsWithRandom: 2,
        epsilon: 0.5,
        epsilonDecay: 0.995,                
        epsilonMin: 0.05,
        gamma: 0.9,                 
    };

    const agentConfig = {
        model: model,
        agentConfig: {
            memorySize: 10000,                      // The size of the agent's memory (Q-Learning)
            batchSize: 256,                        // How many tensors will be given to the network when fit
            temporalWindow: temporalWindow         // The temporal window giving previous inputs & actions
        }
    };

    // First we need an academy to host everything
    const academy = new reimprove.Academy();
    const teacher = academy.addTeacher(teacherConfig);
    const agent = academy.addAgent(agentConfig);

    academy.assignTeacherToAgent(agent, teacher);
    await academy.OnLessonEnded(teacher, async () => await resetGame());

    while(true) {
        // gather results
        const scores = await browser.$$('.score');
        const currentScoreT = await scores[0].getText();
        const bestScoreT = await scores[1].getText();
        let currentScore = parseInt(currentScoreT);
        let bestScore = parseInt(bestScoreT);

        // Gather
        const tiles = await browser.$$('.tile');
        const cellValues = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];
        
        for(let tileI in tiles) {
            let cellIndex = await tiles[tileI].getAttribute("data-index");
            let cellV = await tiles[tileI].getText();
            cellValues[cellIndex] = parseInt(cellV);
        }

        // Step the learning
        let result = await academy.step([{teacherName: teacher, agentsInput: cellValues}]);

        // Take Action
        if(result !== undefined) {
            steps++;
            var action = result.get(agent);
            if(action === 0) {
                browser.keys('Up arrow'); // Up
            } else if(action === 1) {
                browser.keys('Left arrow'); // Left
            } else if(action === 2) {
                browser.keys('Down arrow'); // Down
            } else if(action === 3) {
                browser.keys('Right arrow'); // Right
            }
        }
        await browser.pause(5);

        let newScoreT = await scores[0].getText();
        let newScore = parseInt(newScoreT);

        let reward = (currentScore === newScore)
            ? -0.1
            : newScore - currentScore;
        academy.addRewardToAgent(agent, reward);

        if(currentScore === newScore) {
            stepsWithoutGain++;
        } else {
            stepsWithoutGain = 0;
        }

        console.log("Gen: " + gen + ", Step: " + steps);
        const isFail = await failMessage.isDisplayed();
        if(steps >= 500 || isFail === true || stepsWithoutGain > 20) {
            await resetGame();
            academy.resetTeacherLesson(teacher);
        }
    }
    await browser.deleteSession();
})().catch((e) => console.error(e));
