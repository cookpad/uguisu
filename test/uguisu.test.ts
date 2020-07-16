import { expect as expectCDK, matchTemplate, MatchStyle } from '@aws-cdk/assert';
import * as cdk from '@aws-cdk/core';
import * as Uguisu from '../lib/uguisu-stack';

test('Empty Stack', () => {
    const app = new cdk.App();
    // WHEN
    const stack = new Uguisu.UguisuStack(app, 'MyTestStack');
    // THEN
    expectCDK(stack).to(matchTemplate({
      "Resources": {}
    }, MatchStyle.EXACT))
});
