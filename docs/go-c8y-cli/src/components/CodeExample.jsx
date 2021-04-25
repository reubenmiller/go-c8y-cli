import React from "react"
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@site/src/components/CodeBlock';
// import Highlight, {defaultProps} from 'prism-react-renderer';

const codeTypes = [
    { label: 'Shell', value: 'bash' },
    { label: 'PowerShell', value: 'powershell' },
];

const CodeExample = ({ videoSrcURL, videoTitle, width, height, ...props }) => {

    let firstChild;
    let secondChild;
    
    if (props.children.length > 1) {
        firstChild = props.children[0].props.children.props.children;
        secondChild = props.children[1].props.children.props.children
    } else {
        firstChild = props.children.props.children.props.children
        secondChild = props.children.props.children.props.children
    }

    // console.log('firstChild: ', firstChild);
    // console.log('secondChild: ', secondChild);

    return (
        <Tabs
            groupId="shell-types"
            defaultValue="bash"
            values={codeTypes}
        >
            <TabItem value="bash">
                <CodeBlock render={false} className={"bash"} transform={false}>
                    {firstChild}
                </CodeBlock>
            </TabItem>
            <TabItem value="powershell">
                <CodeBlock render={false} className={"powershell"} transform={true}>
                    {secondChild}
                </CodeBlock>
            </TabItem>
        </Tabs>
    );
};

export default CodeExample;
