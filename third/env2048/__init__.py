from gym.envs.registration import register

register(
    id='env2048-v0',
    entry_point='env2048.env2048:env2048',
)