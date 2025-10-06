import {useState} from 'react';
import './App.css';
import {GenerateCmakeProject, GetExportInfo, Greet, SelectInputFile, SelectOutputDirectory} from "../wailsjs/go/main/App";
import { Button, Input, message, Table, TableColumnsType, Tooltip } from 'antd';
import { TableRowSelection } from 'antd/es/table/interface';
import FunctionTable from './FunctionTable';
interface ExportFunctionInfo {
    name: string;
    rva: number;
    ordinal: number;
}
function App() {
    const [messageApi, contextHolder] = message.useMessage();
    const [dirData, setDirData] = useState<Record<string, string>>({
        // 处理的dll文件
        inputFile: '',
        // cmake项目输出目录
        outputDirectory: '',
    });
    const [exportInfo, setExportInfo] = useState<ExportFunctionInfo[]>([]);
    const [forwardList, setForwardList] = useState<number[]>([]);
    const [generateing, setGenerateing] = useState(false);
    const updateDirData = (k: string, v: string) => setDirData({
        ...dirData,
        [k]: v,
    });
    const getExportInfo = async (filePath: string) => {
        const exportInfo = await GetExportInfo(filePath);
        console.info('exportInfo:', exportInfo)
        const ei = exportInfo.map((ef) => ({
            name: ef.name,
            rva: ef.function_rva,
            ordinal: ef.ordinal,
            forwarder: ef.forwarder,
        }))
        setExportInfo(ei);
    }
    const selectInputFile = async () => {
        const file = await SelectInputFile()
        console.info('file:', file)
        updateDirData('inputFile', file)
        await getExportInfo(file)
    }
    const selectOutputDirectory = async () => {
        const dir = await SelectOutputDirectory()
        if (!dir) {
            messageApi.warning('未选择输出目录')
            return
        }
        console.info('dir:', dir)
        updateDirData('outputDirectory', dir)
    }
    const generateProject = async () => {
        if (!dirData.inputFile || !dirData.outputDirectory) {
            messageApi.warning('请先选择输入文件和输出目录')
            return
        }
        setGenerateing(true)
        const start = messageApi.info('正在生成cmake项目，请稍后...')
        try {
            // 生成cmake项目
            const res = await GenerateCmakeProject(dirData.inputFile, dirData.outputDirectory, forwardList)
            console.info('generate cmake project res:', res)
            start()
            messageApi.success('生成cmake项目成功')
        } catch (e) {
            console.error('generate cmake project error:', e)
            messageApi.error('生成cmake项目失败，请检查输入文件是否正确')
        } finally {
            setGenerateing(false)
        }
    }
    return (
        <div id="app">
            {contextHolder}
            <div className='file-dir-row'>
                <span style={{width: '120px'}}>目标文件：</span>
                <Input value={dirData.inputFile} readOnly />
                &nbsp;
                <Button type='primary' onClick={selectInputFile}>选择</Button>
            </div>
            <div className='file-dir-row'>
                <span style={{width: '120px'}}>输出目录：</span>
                    <Input value={dirData.outputDirectory} readOnly />
                &nbsp;
                <Button type='primary' onClick={selectOutputDirectory}>选择</Button>
            </div>
            <div style={{
                display: 'flex',
                justifyContent: 'right',
                padding: '0 1rem',
            }}>
                <Button type='primary' loading={generateing} onClick={generateProject} disabled={!dirData.inputFile || !dirData.outputDirectory}>生成项目</Button>
            </div>
            <div className='export-table'>
                {/* 函数列表及处理状态 */}
                <FunctionTable exportInfo={exportInfo} onFowardListChange={setForwardList} />
            </div>
        </div>
    )
}

export default App
