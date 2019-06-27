import random
from collections import deque
import numpy as np
import gym
import tensorflow as tf
from keras import backend as K
from keras.layers import Dense
from keras.models import Sequential
from keras.optimizers import Adam
from selenium import webdriver
from selenium.webdriver.common.keys import Keys

class env2048(gym.core.Env):
    metadata = {'render.modes': ['human']}

    def __init__(self):
        # Start browser
        self.driver = webdriver.Firefox()
        #self.driver.get('https://4ark.me/2048/')
        self.driver.get('http://127.0.0.1:8080/')
        self.driver.implicitly_wait(2)
        assert '2048' in self.driver.title
        self.body = self.driver.find_element_by_tag_name('body')
        self.restart_button = self.driver.find_element_by_class_name('restart-btn')
    
    def __del__(self):
        self.driver.close()

    def step(self, action):
        old_state = self.get_state()

        if action == 0:
            self.body.send_keys(Keys.UP)
        elif action == 1:
            self.body.send_keys(Keys.DOWN)
        elif action == 2:
            self.body.send_keys(Keys.LEFT)
        else:
            self.body.send_keys(Keys.RIGHT)

        state = self.get_state()
        reward = self.calculate_reward(state, old_state)
        finished = np.abs(reward) == 1
        return state, reward, finished, self.get_largest_tile(state), {}

    def calculate_added_sum(self, new_state, old_state):
        return sum(new_state) - sum(old_state)

    def count_merges(self, new_state, old_state):
        old_tile_count = np.count_nonzero(old_state)
        new_tile_count = np.count_nonzero(new_state)
        
        if old_tile_count == 16 and new_tile_count == 16:
            return 0
        return old_tile_count - new_tile_count + 1

    def get_largest_tile(self, state):
        return max(state)

    def calculate_reward(self, new_state, old_state):
        if self.did_fail():
            return -1
        
        merges_count = self.count_merges(new_state, old_state)
        if merges_count == 0:
            return -0.1
        if self.get_largest_tile(new_state) == 2048:
            return 1
        return float(merges_count) / 12.0

    def did_fail(self):
        fail_element = self.driver.find_element_by_css_selector('.failure-container')
        return fail_element.get_attribute('class').__contains__('action')

    def get_state(self):
        state = self.driver.execute_script("return data")
        parsed_state = []
        for i in state['cell']:
            parsed_state.append(i['val'])
        return parsed_state

    def reset(self):
        self.restart_button.click()
        return self.get_state()
