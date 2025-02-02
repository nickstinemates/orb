from behave.model_core import Status


def before_scenario(context, scenario):
    context.containers_id = dict()
    context.agent_groups = dict()
    context.existent_sinks_id = list()
    context.tap_tags = dict()


def after_scenario(context, scenario):
    if 'access_denied' in context and context.access_denied is True:
        scenario.set_status(Status.skipped)
    if scenario.status != Status.failed:
        context.execute_steps('''
        Then stop the orb-agent container
        Then remove the orb-agent container
        ''')
    if "driver" in context:
        context.driver.close()
        context.driver.quit()
