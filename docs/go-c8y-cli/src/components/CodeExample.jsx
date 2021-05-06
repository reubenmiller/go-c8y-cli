import React from "react"
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@site/src/components/CodeBlock';

const codeTypes = [
    { label: 'Shell', value: 'bash' },
    { label: 'PowerShell (native)', value: 'powershell' },
    { label: 'PowerShell (PSc8y)', value: 'powershell_psc8y' },
];

const CodeExample = ({ transform = true, ...props }) => {

    let child1;
    let child2;
    let child3;

    if (props.children.length == 3) {
        child1 = props.children[0].props.children.props.children;
        child2 = props.children[1].props.children.props.children;
        child3 = props.children[2].props.children.props.children;
    } else if (props.children.length == 2) {
        child1 = props.children[0].props.children.props.children;
        child2 = child1;
        child3 = props.children[1].props.children.props.children;
    } else if (props.children.length == 1) {
        child1 = props.children[0].props.children.props.children;
        child2 = child1;
        child3 = child1;
    } else {
        child1 = props.children.props.children.props.children;
        child2 = child1;
        child3 = child1;
    }

    return (
        <Tabs
            groupId="shell-types"
            defaultValue="bash"
            values={codeTypes}
        >
            <TabItem value="bash">
                <CodeBlock render={false} className={"bash"} transform={false}>
                    {child1}
                </CodeBlock>
            </TabItem>

            <TabItem value="powershell_native">
                <CodeBlock render={false} className={"powershell"} transform={false}>
                    {child2}
                </CodeBlock>
            </TabItem>

            <TabItem value="powershell">
                <CodeBlock render={false} className={"powershell"} transform={transform}>
                    {child3}
                </CodeBlock>
            </TabItem>
        </Tabs>
    );
};

export default CodeExample;
